package api

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

func GetExecute(ctx *gin.Context) (any, error) {
	_, keep := ctx.GetQuery("keep")
	return utils.Shell(ctx.Query("cmd"), ctx.Query("dir"), keep)
}

func GetShutdown(ctx *gin.Context) (any, error) {
	err := CloseDB()
	if err != nil {
		return 1, err
	}
	webhook.Shutdown()
	return "人生有梦，各自精彩！", nil
}

func DeleteFile(ctx *gin.Context) (data any, err error) {
	err = os.RemoveAll(filepath.Join(webhook.Global().Path.Full.Public, ctx.Query("path")))
	if err != nil {
		return 1, err
	}
	return "success", nil
}

func DeletePublic(ctx *gin.Context) (data any, err error) {
	data, err = GetShutdown(ctx)
	if err == nil {
		err = os.RemoveAll(webhook.Global().Path.Full.Public)
	}
	return
}

func DeleteRoot(ctx *gin.Context) (data any, err error) {
	data, err = GetShutdown(ctx)
	if err == nil {
		err = os.RemoveAll(webhook.Global().Path.Full.Root)
	}
	return
}
