package data

import (
	"os"

	"gorm.io/gorm"
)

var (
	// 全局数据库
	db *gorm.DB
	// 公开文件夹名
	Public string
	// 实际路径
	Resource string
)

func Connect(public, resource string, dialector gorm.Dialector) {
	Public, Resource = public, resource
	// 新建目录
	os.MkdirAll(resource, os.ModePerm)
	db, _ = gorm.Open(dialector, &gorm.Config{})
	// 自动出表
	db.AutoMigrate(new(Attachment))
	db.AutoMigrate(new(Blogger))
	db.AutoMigrate(new(Post))
	db.Table("branches").AutoMigrate(new(Post))
	db.Table("comments").AutoMigrate(new(Post))
}
