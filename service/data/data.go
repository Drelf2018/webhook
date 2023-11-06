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
	Posts    db.DB
	public   string
	folder   string
	replacer *strings.Replacer
)

func Init(r *configs.Config) {
	public = r.Path.Full.Public
	folder = "/" + r.Path.Public
	replacer = strings.NewReplacer(public, "", "\\", "/")
	Posts.SetSqlite(r.Path.Full.Posts).AutoMigrate(&Post{})
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
	err := Posts.DB.Not("local IN ?", files).Find(&a).Error
	if err != nil {
		return err
	}
	asyncio.ForEachPtr(a, func(a *Attachment) { asyncio.RetryError(3, 5, a.Download) })
	return nil
}
