package user

import (
	"database/sql/driver"
	"errors"

	"github.com/glebarez/sqlite"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// 全局数据库
var db *gorm.DB

func init() {
	db, _ = gorm.Open(sqlite.Open("user.db"), &gorm.Config{})
	db.AutoMigrate(new(Jobs))
	db.AutoMigrate(new(User))
}

// 用户
type User struct {
	Uid        string  `gorm:"primaryKey" json:"uid"`
	Token      string  `json:"-"`
	Permission float64 `json:"permission"`
	Jobs       `form:"jobs" json:"jobs" yaml:"jobs"`
	Listening  `form:"listening" json:"listening" yaml:"listening"`
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
