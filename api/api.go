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

// 下载器
var downloader *utils.Downloader

// 请求转发 https://blog.csdn.net/qq_29799655/article/details/113841064
func AnyForwardURL(ctx *gin.Context) (any, error) {
	// 复刻请求
	url := strings.TrimPrefix(ctx.Param("url"), "/")
	req := ctx.Request.Clone(context.Background())
	req.URL.Scheme, url, _ = strings.Cut(url, "/")
	req.URL.Host, url, _ = strings.Cut(url, "/")
	req.URL.Path = "/" + url
	req.Host = req.URL.Host
	req.RemoteAddr = ""
	req.RequestURI = ""
	// 发送新请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1, err
	}
	// 写入状态码和 Header
	ctx.Status(resp.StatusCode)
	header := ctx.Writer.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	// 将 Body 写回原请求中
	_, err = io.Copy(ctx.Writer, resp.Body)
	if err != nil {
		return 2, err
	}
	return nil, nil
}

var visitor = group.Group{
	Handlers: group.H{
		GetVersion,
		GetOnline,
		PostRegister,
		GetToken,
		GetUserUID,
		PostFilters,
		GetTasks,
		GetBlogs,
		GetBlogID,
	},
}

var user = group.Group{
	Middlewares: group.M{IsUser},
	Handlers: group.H{
		GetFollowing,
		PostBlog,
		PostTask,
		GetTaskID,
		PatchTaskID,
		DeleteTaskID,
		GetUser,
		PostTest,
		PostTests,
		PatchUserUID,
	},
}

var admin = group.Group{
	Middlewares: group.M{IsAdmin},
	CustomFunc:  func(r gin.IRouter) { r.StaticFS("/logs", http.Dir(config.Path.Full.Logs)) },
}

var owner = group.Group{
	Middlewares: group.M{IsOwner},
	CustomFunc:  func(r gin.IRouter) { r.StaticFS("/root", http.Dir(config.Path.Full.Root)) },
	Handlers: group.H{
		GetExecute,
		GetShutdown,
		PostUpload,
	},
}

var api = group.Group{
	Path:      "api",
	Handlers:  group.H{GetValid, GetPing},
	Groups:    group.G{visitor, user, admin, owner},
	Convertor: group.Convertor,
}

var Backend = group.Group{
	Middlewares: group.M{group.CORS, Index},
	CustomFunc: func(r gin.IRouter) {
		downloader = utils.NewDownloader(config.Path.Full.Public)
		r.StaticFS("/public", downloader)
	},
	HandlerMap: map[string]group.HandlerFunc{
		"/forward/*url": AnyForwardURL,
	},
	Groups: group.G{api},
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
