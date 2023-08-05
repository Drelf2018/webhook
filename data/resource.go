package data

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	// 全局数据库
	db *gorm.DB
	// 资源文件夹
	Resource string
	// 数据库文件名
	File string
)

func Connect(resource, file string) {
	Resource, File = resource, file
	// 新建目录
	os.MkdirAll(Resource, os.ModePerm)
	db, _ = gorm.Open(sqlite.Open(filepath.Join(Resource, File)), &gorm.Config{})
	// 自动出表
	db.AutoMigrate(new(Attachment))
	db.AutoMigrate(new(Blogger))
	db.AutoMigrate(new(Post))
	db.Table("branches").AutoMigrate(new(Post))
	db.Table("comments").AutoMigrate(new(Post))
}
