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
		fs.Register(file.WeiboClient{})
		r.StaticFS("/public", fs)
	},
	Handlers: []group.H{
		GetVersion,
		GetValid,
		GetOnline,
		PostRegister,
		GetToken,
		GetUUID,
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
		GetExec,
		GetUserUID,
		GetShutdown,
		DeletePublic,
		DeleteFile,
		DeleteRoot,
	},
}

var api = group.G{
	Middleware: group.CORS,
	Handlers:   []group.H{GetPing},
	Groups:     []group.G{vistor, user, admin, owner},
}

func New() (r *gin.Engine) {
	return api.New()
}

func Default() (r *gin.Engine) {
	return api.Default()
}
