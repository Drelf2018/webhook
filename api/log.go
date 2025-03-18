package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "unsafe"
)

//go:linkname logger
var logger *logrus.Logger

func Info(ctx *gin.Context) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		logger.Infof(`%s %s "%s" (%s)`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, value)
	} else {
		logger.Infof(`%s %s "%s"`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL)
	}
}

func Error(ctx *gin.Context, err error) {
	if value, exists := ctx.Get(MagicUIDKey); exists {
		logger.Errorf(`%s %s "%s": %s (%s)`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, err, value)
	} else {
		logger.Errorf(`%s %s "%s": %s`, ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL, err)
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
