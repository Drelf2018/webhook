package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	if stop != nil {
		stop()
	}
	if cancel != nil {
		cancel()
	}
	return "人生有梦，各自精彩！", nil
}

func GetUserUID(ctx *gin.Context) (any, error) {
	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB.Preload("Tasks").Preload("Tasks.Filters").Preload("Tasks.Logs").Limit(1).Find(user)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrUserNotExist
	}
	return user, nil
}

// 上传文件
func PostUpload(ctx *gin.Context) (any, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return 1, err
	}
	errs := make(utils.JoinError, 0)
	upload := config.Path.Full.Upload
	for fieldname, files := range form.File {
		dir := filepath.Join(form.Value[fieldname]...)
		if strings.HasPrefix(dir, "user") || strings.HasPrefix(dir, "admin") || strings.HasPrefix(dir, "owner") {
			errs = append(errs, fmt.Errorf("dir \"%s\" has invalid prefix", dir))
			continue
		}
		for _, file := range files {
			if file.Filename == "index.html" {
				file.Filename = time.Now().Format("index.2006-01-02-15-04-05.html")
			}
			filename := filepath.Join(upload, dir, file.Filename)
			err := ctx.SaveUploadedFile(file, filename)
			if err != nil {
				errs = append(errs, err)
			}
			LoadFile(upload, filename)
		}
	}
	if len(errs) != 0 {
		return 2, fmt.Errorf("webhook/api: upload files error: %w", errs)
	}
	return "success", nil
}
