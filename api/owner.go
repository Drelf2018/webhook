package api

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

func GetExec(ctx *gin.Context) (any, error) {
	return utils.Shell(ctx.Query("cmd"), ctx.Query("dir"))
}

func GetUserUID(ctx *gin.Context) (any, error) {
	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB().Preload("Tasks").Limit(1).Find(user)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrUserNotExist
	}
	return user, nil
}

func GetShutdown(ctx *gin.Context) (any, error) {
	err := CloseDB()
	if err != nil {
		return 1, err
	}
	webhook.Shutdown()
	return "人生有梦，各自精彩！", nil
}

func DeletePublic(ctx *gin.Context) (data any, err error) {
	data, err = GetShutdown(ctx)
	if err == nil {
		os.RemoveAll(webhook.Global().Path.Full.Public)
	}
	return
}

func DeleteFile(ctx *gin.Context) (data any, err error) {
	err = os.RemoveAll(filepath.Join(webhook.Global().Path.Full.Public, ctx.Query("path")))
	if err != nil {
		return 1, err
	}
	return "success", nil
}
