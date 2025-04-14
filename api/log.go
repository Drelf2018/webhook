package api

import (
	"net/http"

	group "github.com/Drelf2018/gin-group"
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

func init() {
	group.DefaultConvertor = func(f group.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			if data, err := f(ctx); err == nil {
				Info(ctx)
				if data != nil {
					ctx.JSON(http.StatusOK, group.Response{Code: 0, Error: "", Data: data})
				}
			} else {
				Error(ctx, err)
				if code, ok := data.(int); ok {
					ctx.JSON(http.StatusOK, group.Response{Code: code, Error: err.Error(), Data: nil})
				} else {
					ctx.JSON(http.StatusOK, group.Response{Code: 1, Error: err.Error(), Data: data})
				}
			}
		}
	}
}
