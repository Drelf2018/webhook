package data

import (
	"io/fs"
	"path/filepath"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/db"
)

var (
	Posts  db.DB
	public string
	folder string
)

func initTestPost() {
	TestPost.Blogger.Face.Init()
	TestPost.Repost.Blogger.Face.Init()
	TestPost.Repost.Attachments[0].Init()
}

func Init(r *configs.Config) {
	public = r.Path.Full.Public
	folder = "/" + r.Path.Public
	initTestPost()
	Posts.SetSqlite(r.Path.Full.Posts).AutoMigrate(&Post{})
}

func CheckFiles() error {
	files := make([]string, 0)
	filepath.Walk(public, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	var as Attachments
	err := Posts.DB.Not("local IN ?", files).Find(&as).Error
	if err != nil {
		return err
	}

	asyncio.ForEachPtr(as, func(a *Attachment) { asyncio.RetryError(3, 5, a.Download) })
	return nil
}
