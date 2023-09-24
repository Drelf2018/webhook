package data

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/resource"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/db"
)

var (
	Data   db.DB
	folder string
	public resource.Explorer
)

func Init(r *configs.Config) {
	folder = "/" + r.Path.Public
	public = r.Resource.MakeTo(r.Path.Public)
	public.MkdirAll()
	Data.SetSqlite(public.Path(r.Path.Posts)).AutoMigrate(&Post{})
}

func Public() resource.Explorer {
	return public
}

func CheckFiles() error {
	files := make([]string, 0)
	rep := strings.NewReplacer(public.Path(), "", "\\", "/")
	filepath.Walk(public.Path(), func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, rep.Replace(path))
		}
		return nil
	})
	var a Attachments
	err := Data.DB.Not("local IN ?", files).Find(&a).Error
	if err != nil {
		return err
	}
	asyncio.ForEach(a, func(a Attachment) { asyncio.RetryError(3, 5, a.Download) })
	return nil
}
