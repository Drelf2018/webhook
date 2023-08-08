package data

import (
	"os"

	"gorm.io/gorm"
)

var (
	// 全局数据库
	db *gorm.DB
	// 资源路径
	Path   string
	Public string
)

func Connect(path, public string, dialector gorm.Dialector) {
	Path, Public = path, public
	// 新建目录
	os.MkdirAll(path, os.ModePerm)
	db, _ = gorm.Open(dialector, &gorm.Config{})
	// 自动出表
	db.AutoMigrate(new(Attachment))
	db.AutoMigrate(new(Blogger))
	db.AutoMigrate(new(Post))
	db.Table("branches").AutoMigrate(new(Post))
}
