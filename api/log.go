package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Info(ctx *gin.Context) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		Log.Infof(`%s %s "%s" (%s)`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, value)
	} else {
		Log.Infof(`%s %s "%s"`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL)
	}
}

func Error(ctx *gin.Context, err error) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		Log.Errorf(`%s %s "%s": %s (%s)`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, err, value)
	} else {
		Log.Errorf(`%s %s "%s": %s`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, err)
	}
}

func LogMiddleware(ctx *gin.Context) {
	ctx.Next()
	if len(ctx.Errors) != 0 {
		Error(ctx, ctx.Errors.Last())
		ctx.Errors = nil
	} else {
		Info(ctx)
	}
}
