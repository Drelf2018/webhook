package api

import (
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
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

func GetUserUID(ctx *gin.Context) (any, error) {
	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB().Preload("Tasks").Preload("Tasks.Filters").Preload("Tasks.Logs").Limit(1).Find(user)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrUserNotExist
	}
	return user, nil
}
