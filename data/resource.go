package data

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Drelf2018/webhook/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	// 全局数据库
	db *gorm.DB
	// 资源文件夹
	Resource = "resource"
	// 数据库文件名
	File = "posts.db"
)

func __init__() {
	// 记录本次参数
	os.WriteFile(".storeage", []byte(fmt.Sprintf("\"%s\".\"%s\"", Resource, File)), os.ModePerm)
	utils.HideFile(".storeage")
	// 新建目录
	os.MkdirAll(Resource, os.ModePerm)
	db, _ = gorm.Open(sqlite.Open(filepath.Join(Resource, File)), &gorm.Config{})
	// 自动出表
	db.AutoMigrate(new(Attachment))
	db.AutoMigrate(new(Blogger))
	db.AutoMigrate(new(Post))
	db.Table("branches").AutoMigrate(new(Post))
}

func init() {
	b, err := os.ReadFile(".storeage")
	if err == nil {
		sList := strings.Split(string(b), "\".\"")
		Resource = sList[0][1:]
		File = sList[1][:len(sList[1])-1]
	}
	__init__()
}

// 重设参数
func Reset(resource, file string) {
	// 从 .storeage 中读到了上次重设的参数
	if Resource == resource && File == file {
		return
	}
	// 断开连接
	sqlDB, _ := db.DB()
	sqlDB.Close()
	// 移除原资源文件夹
	os.RemoveAll(Resource)
	// 重新初始化
	Resource, File = resource, file
	__init__()
}
