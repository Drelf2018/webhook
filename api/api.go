package api

import (
	"net/http"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/file"
	"github.com/gin-gonic/gin"
)

var vistor = group.G{
	Middleware: LogMiddleware,
	CustomFunc: func(r gin.IRouter) {
		fs := file.NewDownloader(webhook.Global().Path.Full.Public)
		fileServer := http.StripPrefix("/public", http.FileServer(fs))
		handler := func(c *gin.Context) {
			if c.Request.URL.RawQuery != "" {
				c.Request.URL.Path = c.Request.URL.Path + "?" + c.Request.URL.RawQuery
				c.Request.URL.RawQuery = ""
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
		r.GET("/public/*filepath", handler)
		r.HEAD("/public/*filepath", handler)
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
	},
}

var api = group.G{
	Middleware: group.CORS,
	Handlers:   []group.H{GetValid, GetPing},
	Groups:     []group.G{vistor, user, admin, owner},
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
	return nil
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
