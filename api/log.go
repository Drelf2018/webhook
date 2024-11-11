package api

import (
	"path/filepath"
	"time"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
)

var log *logrus.Logger

func Log() *logrus.Logger {
	if log == nil {
		hook := &utils.DateHook{Format: filepath.Join(webhook.Global().Path.Full.Logs, "2006-01-02.log")}
		log = &logrus.Logger{
			Out:   hook,
			Hooks: make(logrus.LevelHooks),
			Formatter: &nested.Formatter{
				HideKeys:        true,
				NoColors:        true,
				TimestampFormat: time.TimeOnly,
				ShowFullLevel:   true,
			},
			Level: logrus.DebugLevel,
		}
		log.AddHook(hook)
	}
	return log
}

func Info(ctx *gin.Context) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		Log().Infof(`%s %s "%s" (%s)`, ctx.RemoteIP(), ctx.Request.Method, ctx.Request.URL, value)
	} else {
		Log().Infof(`%s %s "%s"`, ctx.RemoteIP(), ctx.Request.Method, ctx.Request.URL)
	}
}

func Error(ctx *gin.Context, err error) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		Log().Errorf(`%s %s "%s": %s (%s)`, ctx.RemoteIP(), ctx.Request.Method, ctx.Request.URL, err, value)
	} else {
		Log().Errorf(`%s %s "%s": %s`, ctx.RemoteIP(), ctx.Request.Method, ctx.Request.URL, err)
	}
}

func LogMiddleware(ctx *gin.Context) {
	ctx.Next()
	if len(ctx.Errors) != 0 {
		// ctx.Abort()
		// c.JSON(code, jsonObj)
		Error(ctx, ctx.Errors.Last())
		ctx.Errors = nil
	} else {
		Info(ctx)
	}
}

// func LogUIDMiddleware(ctx *gin.Context) {
// 	ctx.Next()
// 	if len(ctx.Errors) == 0 {
// 		InfoUID(ctx, GetUID(ctx))
// 		return
// 	}
// 	ErrorUID(ctx, ctx.Errors.Last(), GetUID(ctx))
// 	ctx.Errors = nil
// }
