package data

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/db"
)

var (
	Data     db.DB
	public   string
	folder   string
	replacer *strings.Replacer
)

func Init(r *configs.Config) {
	folder = "/" + r.Path.Public
	public = r.Path.Full.Public
	replacer = strings.NewReplacer(public, "", "\\", "/")
	Data.SetSqlite(r.Path.Full.Posts).AutoMigrate(&Post{})
}

func CheckFiles() error {
	files := make([]string, 0)
	filepath.Walk(public, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, replacer.Replace(path))
		}
		return nil
	})
	var a Attachments
	err := Data.DB.Not("local IN ?", files).Find(&a).Error
	if err != nil {
		return err
	}
	asyncio.ForEachPtr(a, func(a *Attachment) { asyncio.RetryError(3, 5, a.Download) })
	return nil
}
