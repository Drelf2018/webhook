package model

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/webhook/config"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

var userDB = gorms.SetSQLite("user.db").AutoMigrate(&Job{}, &User{})
var ErrNoSubmitter = errors.New("the post has no submitter")

// 提交者
type User struct {
	Permission `json:"permission"`
	// b站序号
	Uid string `json:"uid" gorm:"primaryKey"`
	// 鉴权码
	Auth string `json:"auth,omitempty"`
	// 经验
	Exp int64 `json:"exp"`
	// 账户下任务
	Jobs []Job `json:"jobs,omitempty" gorm:"foreignKey:UserUid"`
	// 关注的账号
	Follow pq.StringArray `gorm:"type:text[]" json:"follow,omitempty"`
}

func (u *User) Level() int {
	switch {
	case u.Exp < 200:
		return 1
	case u.Exp < 1500:
		return 2
	case u.Exp < 4500:
		return 3
	case u.Exp < 10800:
		return 4
	case u.Exp < 28800:
		return 5
	default:
		return 6
	}
}

func (u *User) String() string {
	return fmt.Sprintf("%s@LV%d", u.Uid, u.Level())
}

func (u *User) RemoveJobs(jobs []string) error {
	return userDB.Debug().Delete(&Job{UserUid: u.Uid}, jobs).Error
}

func (u *User) Update() error {
	return userDB.Updates(u).Error
}

func (u *User) Scan(val any) error {
	return userDB.Select("uid", "permission").First(u, "uid = ?", val).Error()
}

func (u *User) Value() (driver.Value, error) {
	if u == nil {
		return 0, ErrNoSubmitter
	}
	return u.Uid, nil
}

// 构造函数
func NewUser(uid string) (u *User) {
	u = &User{
		Uid:  uid,
		Auth: uuid.NewV4().String(),
	}
	// permission
	switch {
	case uid == config.Global.Owner:
		u.Permission = Owner
	case slices.Contains(config.Global.Administrators, uid):
		u.Permission = Administrator
	case slices.Contains(config.Global.Trustors, uid):
		u.Permission = Trustor
	}
	// save
	userDB.Create(u)
	return
}

// 根据 auth 查询
func QueryUser(auth string) *User {
	if auth == "" {
		return nil
	}
	u := &User{Auth: auth}
	if userDB.PreloadOK(u) {
		return u
	}
	return nil
}

// 根据 uid 查询
func QueryAuth(uid string) *User {
	if uid == "" {
		return nil
	}
	u := new(User)
	if userDB.FirstOK(u, "uid = ?", uid) {
		return u
	}
	return nil
}

func Users(conds ...any) (users []User, err error) {
	err = userDB.Clone().Preloads(&users, conds...).Error()
	return
}

var ErrNoOwner = errors.New("owner permissions cannot be revoked")

// 修改权限
func UpdatePermission(uid string, p Permission) error {
	if config.Global.Owner == uid && !p.Has(Owner) {
		return ErrNoOwner
	}
	return userDB.Updates(User{Uid: uid, Permission: p}).Error
}
