package api

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/utils"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var downloader *utils.Downloader

// 请求转发 https://blog.csdn.net/qq_29799655/article/details/113841064
func ForwardURL(ctx *gin.Context) {
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
}

var vistor = group.G{
	Middleware: LogMiddleware,
	CustomFunc: func(r gin.IRouter) {
		// 请求转发
		r.Any("/forward/*url", ForwardURL)

		// 下载文件
		downloader = utils.NewDownloader(config.Path.Full.Public)
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
		GetFollowing,
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

var API = group.G{
	Middlewares: []gin.HandlerFunc{group.CORS, Index},
	Handlers:    []group.H{GetValid, GetPing},
	Groups:      []group.G{vistor, user, admin, owner},
}

func Initial(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{Filename: "config.toml"}
	}
	if cfg.Role.Admin == nil {
		cfg.Role.Admin = make([]string, 0)
	}
	if cfg.Extra == nil {
		cfg.Extra = make(map[string]any)
	}

	err := cfg.Import()
	if err != nil {
		return err
	}
	err = initial.Initial(cfg)
	if err != nil {
		return err
	}

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

	err = cfg.Export()
	if err != nil {
		return err
	}

	switch cfg.Server.Mode {
	case gin.ReleaseMode, gin.DebugMode, gin.TestMode:
		gin.SetMode(cfg.Server.Mode)
	}

	if logger == nil {
		hook := &utils.DateHook{Format: filepath.Join(cfg.Path.Full.Logs, "2006-01-02.log")}
		logger = &logrus.Logger{
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
		logger.AddHook(hook)
	}

	if AutoSave {
		stop = time.AfterFunc(utils.NextTimeDuration(4, 0, 0), func() {
			cfg.Path.CopyBlogDB()
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-running.Done():
					return
				case <-ticker.C:
					go cfg.Path.CopyBlogDB()
				}
			}
		}).Stop
	}

	config = cfg

	return nil
}
