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
	"golang.org/x/exp/slices"
)

// 全局数据库
var Users db.DB

func Init(r *configs.Config) {
	Users.SetSqlite(r.Path.Full.Users).AutoMigrate(&Job{}, &User{})
	bili.Data["oid"] = r.Oid
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
	Uid        string    `json:"uid" gorm:"primaryKey"`
	Token      string    `json:"token,omitempty"`
	Permission float64   `json:"permission"`
	Jobs       []Job     `json:"jobs,omitempty" gorm:"references:Uid"`
	Listening  Listening `json:"listening,omitempty"`
}

func (u *User) String() string {
	return fmt.Sprintf("User(%v, %v)", u.Uid, u.Permission)
}

func (u *User) RemoveJobs(jobs []string) error {
	defer Users.Preload(u)
	return Users.DB.Model(&Job{}).Where("user_uid = ? and id IN ?", u.Uid, jobs).Update("user_uid", nil).Error
}

func (u *User) Update() error {
	return Users.DB.Updates(u).Error
}

// 检查回复
func (u *User) MatchReplies() (bool, error) {
	replies, err := GetReplies()
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(replies, func(r *Replie) bool {
		return r.Member.Mid == u.Uid && r.Content.Message == u.Token
	}), nil
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
	return Users.Select(u, []string{"uid", "permission"}, "uid = ?", val).Error()
}

func (u User) Value() (driver.Value, error) {
	return u.Uid, nil
}

// 构造函数
func Make(uid string) *User {
	u := User{
		Uid:        uid,
		Token:      uuid.NewV4().String(),
		Permission: 5.10,
	}
	if slices.Contains(configs.Get().Administrators, uid) {
		u.Permission = 1.0
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
