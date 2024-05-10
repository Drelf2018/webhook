package api

import (
	"os"
	"runtime"
	"time"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/config"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func WrapError(data any, err error) (any, group.Error) {
	if err == nil {
		return data, nil
	}
	_, _, line, _ := runtime.Caller(1)
	return data, group.NewError(line, err.Error())
}

var Api = group.Group{
	Middlewares: gin.HandlersChain{group.CORS},
	Customize: func(r gin.IRouter) {
		r.StaticFS("/public", request.DefaultDownloadSystem(config.Global.Path.FullPath.Public))
	},
	Groups: []group.Group{Visitor, {
		Middlewares: gin.HandlersChain{SetUser},
		Groups:      []group.Group{Submitter, Admin},
	}},
}

func GetEngine() (*gin.Engine, error) {
	file, err := os.OpenFile(config.Global.Path.FullPath.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	log = &logrus.Logger{
		Out: file,
		Formatter: &nested.Formatter{
			HideKeys:        true,
			NoColors:        true,
			TimestampFormat: time.DateTime,
		},
		Level: logrus.DebugLevel,
	}

	Api.Middlewares = append(Api.Middlewares, group.Static(config.Global.Path.FullPath.Views))
	return group.Default(Api), nil
}
