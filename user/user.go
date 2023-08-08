package user

import (
	"database/sql/driver"
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

var (
	// 全局数据库
	db *gorm.DB
	// 动态评论区获取链接
	Url string
)

func Connect(oid string, dialector gorm.Dialector) {
	Url = fmt.Sprintf("https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies", oid)
	db, _ = gorm.Open(dialector, &gorm.Config{})
	db.AutoMigrate(new(Jobs))
	db.AutoMigrate(new(User))
}

// 用户
type User struct {
	Uid        string  `gorm:"primaryKey" json:"uid"`
	Token      string  `json:"-"`
	Permission float64 `json:"permission"`
	Jobs       `form:"jobs" json:"-"`
	Listening  `form:"listening" json:"-"`
}

// 构造函数
func (u *User) Make(uid string) *User {
	u.Uid = uid
	u.Token = uuid.NewV4().String()
	u.Permission = 5.10
	db.Create(u)
	return u
}

// 根据 uid 查询
func (u *User) Query(token string) *User {
	if token == "" {
		return nil
	}
	result := db.First(u, "token = ?", token)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return u
}

// 升级
func (u *User) LevelUP() {
	u.Permission -= 0.01
	switch u.Permission {
	case 5.0:
		u.Permission = 4.25
	case 4.0:
		u.Permission = 3.50
	case 3.0:
		u.Permission = 2.75
	case 0.99:
		u.Permission = 1.0
	}
	db.Model(new(User)).Where("uid = ?", u.Uid).Update("permission", u.Permission)
}

func (u *User) Scan(val any) error {
	return db.First(u, "uid = ?", val).Error
}

func (u User) Value() (driver.Value, error) {
	return u.Uid, nil
}
