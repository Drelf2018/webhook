package user

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/db"
	"github.com/Drelf2020/utils"
	uuid "github.com/satori/go.uuid"
)

// 全局数据库
var Users db.DB

func Init(r *configs.Config) {
	Users.SetSqlite(r.Path.Full.Users)
	Users.AutoMigrate(&Job{})
	Users.AutoMigrate(&User{})
	SetApi(r.Oid)
}

// 监听列表获取错误
var ErrNotListeningList = errors.New("不是一个好的监听列表")

// 监听列表的读取实现
type Listening []string

func (l *Listening) Scan(val any) error {
	if val, ok := val.(string); ok {
		*l = utils.Ternary(val == "", []string{}, strings.Split(val, ","))
		return nil
	}
	return ErrNotListeningList
}

func (l Listening) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}

// 用户
type User struct {
	Uid        string    `gorm:"primaryKey" json:"uid"`
	Token      string    `json:"-"`
	Permission float64   `json:"permission"`
	Jobs       []Job     `gorm:"references:Uid" form:"jobs" json:"-"`
	Listening  Listening `form:"listening" json:"-"`
}

func (u *User) String() string {
	return fmt.Sprintf("User(%v, %v)", u.Uid, u.Permission)
}

func (u *User) RemoveJobs(jobs []string) error {
	defer Users.First(u)
	return Users.DB.Model(&Job{}).Where("user_uid = ? and id IN ?", u.Uid, jobs).Update("user_uid", nil).Error
}

func (u *User) Update() error {
	return Users.DB.Updates(u).Error
}

// 构造函数
func Make(uid string) *User {
	u := User{
		Uid:        uid,
		Token:      uuid.NewV4().String(),
		Permission: 5.10,
	}
	Users.DB.Create(&u)
	return &u
}

// 根据 uid 查询
func Query(token string) *User {
	if token == "" {
		return nil
	}
	var u User
	if Users.Preload(&u, "token = ?", token).NoRecord() {
		return nil
	}
	return &u
}

// 修改权限
func (u User) UpdatePermission() error {
	return Users.DB.Model(u).Update("permission", u.Permission).Error
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
	u.UpdatePermission()
}

func (u *User) Scan(val any) error {
	return Users.Base(u, "uid = ?", val).Error()
}

func (u User) Value() (driver.Value, error) {
	return u.Uid, nil
}
