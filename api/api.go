package api

import (
	"context"
	"io"
	"net/http"
	"strings"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/file"
	"github.com/gin-gonic/gin"
)

var downloader *file.Downloader

var vistor = group.G{
	Middleware: LogMiddleware,
	CustomFunc: func(r gin.IRouter) {
		// 下载文件的处理函数
		downloader = file.NewDownloader(webhook.Global().Path.Full.Public)
		fileServer := http.StripPrefix("/public", http.FileServer(downloader))
		publicHandler := func(c *gin.Context) {
			if c.Request.URL.RawQuery != "" {
				c.Request.URL.Path = c.Request.URL.Path + "?" + c.Request.URL.RawQuery
				c.Request.URL.RawQuery = ""
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
		r.GET("/public/*filepath", publicHandler)
		r.HEAD("/public/*filepath", publicHandler)

		// 请求转发的处理函数 https://blog.csdn.net/qq_29799655/article/details/113841064
		r.Any("/forward/*url", func(ctx *gin.Context) {
			// 复刻请求
			req := ctx.Request.Clone(context.Background())
			url := strings.Replace(req.URL.Path, "/forward/", "", 1)
			req.URL.Scheme, url, _ = strings.Cut(url, "/")
			req.URL.Host, url, _ = strings.Cut(url, "/")
			req.URL.Path = "/" + url
			req.Host = req.URL.Host
			req.RemoteAddr = ""
			req.RequestURI = ""
			// 发送新请求
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				ctx.JSON(http.StatusOK, group.Response{Code: 1, Error: err.Error()})
				return
			}
			// 写入状态码和 Header
			ctx.Status(resp.StatusCode)
			header := ctx.Writer.Header()
			for k, vs := range resp.Header {
				for _, v := range vs {
					header.Add(k, v)
				}
			}
			// 将 Body 写入原请求中
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				ctx.JSON(http.StatusOK, group.Response{Code: 2, Error: err.Error()})
				return
			}
			ctx.Writer.Write(b)
		})
	},
	Handlers: []group.H{
		GetVersion,
		GetOnline,
		PostRegister,
		GetToken,
		GetUUID,
		PostFilter,
		GetBlogs,
		GetBlogID,
	},
}

var user = group.G{
	Path:       "user",
	Middleware: IsUser,
	Handlers: []group.H{
		PostBlog,
		PostTask,
		GetTaskID,
		PatchTaskID,
		DeleteTaskID,
		Get,
		group.Wrapper(http.MethodPatch, "/:uid", PatchUser),
		PostTest,
		PostTests,
	},
}

var admin = group.G{
	Path:       "admin",
	Middleware: IsAdmin,
	CustomFunc: func(r gin.IRouter) { r.StaticFS("/logs", http.Dir(webhook.Global().Path.Full.Logs)) },
}

var owner = group.G{
	Path:       "owner",
	Middleware: IsOwner,
	CustomFunc: func(r gin.IRouter) { r.StaticFS("/root", http.Dir(webhook.Global().Path.Full.Root)) },
	Handlers: []group.H{
		GetExecute,
		GetShutdown,
		GetUserUID,
		PostUpload,
	},
}

var api = group.G{
	Middlewares: []gin.HandlerFunc{group.CORS, Index}, //
	Handlers:    []group.H{GetValid, GetPing},
	Groups:      []group.G{vistor, user, admin, owner},
}

func load() error {
	var users []UserClaims
	err := UserDB().Find(&users).Error
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.IssuedAt != 0 {
			tokenIssuedAt.Store(user.UID, user.IssuedAt)
		}
	}

	upload := webhook.Global().Path.Full.Upload
	return LoadDir(upload, upload)
}

func New() (r *gin.Engine) {
	err := load()
	if err != nil {
		panic(err)
	}
	return api.New()
}

func Default() (r *gin.Engine) {
	err := load()
	if err != nil {
		panic(err)
	}
	return api.Default()
}
