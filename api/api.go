package api

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/file"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/Drelf2018/webhook/utils"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var downloader *file.Downloader

var vistor = group.G{
	Middleware: LogMiddleware,
	CustomFunc: func(r gin.IRouter) {
		// 下载文件的处理函数
		downloader = file.NewDownloader(config.Path.Full.Public)
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
		GetRssID,
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
	CustomFunc: func(r gin.IRouter) { r.StaticFS("/logs", http.Dir(config.Path.Full.Logs)) },
}

var owner = group.G{
	Path:       "owner",
	Middleware: IsOwner,
	CustomFunc: func(r gin.IRouter) { r.StaticFS("/root", http.Dir(config.Path.Full.Root)) },
	Handlers: []group.H{
		GetExecute,
		GetShutdown,
		GetUserUID,
		PostUpload,
	},
}

var api = group.G{
	Middlewares: []gin.HandlerFunc{group.CORS, Index},
	Handlers:    []group.H{GetValid, GetPing},
	Groups:      []group.G{vistor, user, admin, owner},
}

var config *webhook.Config

type OpenAPI struct {
	*gin.Engine
}

func (o *OpenAPI) Initial(cfg *webhook.Config) error {
	if o == nil || o.Engine == nil {
		return ErrOpenAPINotExist
	}
	config = cfg

	if BaseURL == "" {
		baseURL, ok := cfg.Extra["base_url"]
		if ok {
			BaseURL, _ = baseURL.(string)
		} else {
			cfg.Extra["base_url"] = ""
		}
	}

	if !AutoDownload {
		autoDownload, ok := cfg.Extra["auto_download"]
		if ok {
			AutoDownload, _ = autoDownload.(bool)
		} else {
			cfg.Extra["auto_download"] = false
		}
	}

	if JWTSecretKey == nil {
		jwt, ok := cfg.Extra["jwt_secret_key"]
		if ok {
			s, _ := jwt.(string)
			JWTSecretKey = []byte(s)
		} else {
			JWTSecretKey = []byte("my_secret_key")
			cfg.Extra["jwt_secret_key"] = string(JWTSecretKey)
		}
	}

	if Log == nil {
		hook := &utils.DateHook{Format: filepath.Join(config.Path.Full.Logs, "2006-01-02.log")}
		Log = &logrus.Logger{
			Out:   hook,
			Hooks: make(logrus.LevelHooks),
			Formatter: &nested.Formatter{
				HideKeys:        true,
				NoColors:        true,
				TimestampFormat: "15:04:05",
				ShowFullLevel:   true,
			},
			Level: logrus.DebugLevel,
		}
		Log.AddHook(hook)
	}

	var err error
	if UserDB == nil {
		UserDB, err = gorm.Open(sqlite.Open(cfg.Path.Full.UserDB))
		if err != nil {
			return err
		}
		err = UserDB.AutoMigrate(&model.User{}, &model.Task{}, &model.Filter{}, &model.RequestLog{})
		if err != nil {
			return err
		}
	}

	if BlogDB == nil {
		BlogDB, err = gorm.Open(sqlite.Open(cfg.Path.Full.BlogDB))
		if err != nil {
			return err
		}
		err = BlogDB.AutoMigrate(&model.Blog{})
		if err != nil {
			return err
		}
	}

	var users []UserClaims
	err = UserDB.Find(&users).Error
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.IssuedAt != 0 {
			tokenIssuedAt.Store(user.UID, user.IssuedAt)
		}
	}

	err = LoadDir(cfg.Path.Full.Upload, cfg.Path.Full.Upload)
	if err != nil {
		return err
	}

	err = registrar.Initial(cfg.Extra)
	if err != nil {
		return err
	}

	err = cfg.Export()
	if err != nil {
		return err
	}

	api.Bind(o)
	return nil
}
