package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

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
		if strings.HasPrefix(dir, "api") {
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
	return Success, nil
}

// 执行命令
func GetExecute(ctx *gin.Context) (any, error) {
	_, keep := ctx.GetQuery("keep")
	return utils.Shell(ctx.Query("cmd"), ctx.Query("dir"), keep)
}

// 优雅关机
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
