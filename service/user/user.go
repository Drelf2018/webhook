package user

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Drelf2018/request"
	"github.com/Drelf2020/utils"
	"github.com/glebarez/sqlite"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// 全局数据库
	db *gorm.DB
	// 动态评论区获取链接
	url string
)

func SetDB(r *gorm.DB) *gorm.DB {
	db = r
	db.AutoMigrate(new(User))
	return db
}

func SetDialector(dialector gorm.Dialector) *gorm.DB {
	db, _ = gorm.Open(dialector, &gorm.Config{})
	return SetDB(db)
}

func SetSqlite(file string) *gorm.DB {
	return SetDialector(sqlite.Open(file))
}

func SetOid(oid string) {
	url = fmt.Sprintf("https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies", oid)
}

func Exists[T any](conds ...any) bool {
	return !errors.Is(db.Preload(clause.Associations).First(new(T), conds...).Error, gorm.ErrRecordNotFound)
}

func Update(x any, conds ...any) bool {
	return !errors.Is(db.Preload(clause.Associations).First(x, conds...).Error, gorm.ErrRecordNotFound)
}

func CreateTestUser() {
	if Exists[User]("uid = ?", "188888131") {
		return
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Create(&User{
		Uid:        "188888131",
		Token:      "********",
		Permission: 1,
		Jobs: []Job{
			{
				Patten: "bilibili434334701",
				Job: request.Job{
					Method: http.MethodPost,
					Url:    "https://postman-echo.com/post",
					Data: request.M{
						"msg":    "{content}",
						"uid":    "{uid}",
						"origin": "{text}",
					},
				},
			},
		},
		Listening: make(Listening, 0),
	})
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

func (u *User) RemoveJobs(jobs []string) error {
	defer Update(u)
	return db.Model(&Job{}).Where("user_uid = ? and id IN ?", u.Uid, jobs).Update("user_uid", nil).Error
}

func (u *User) Update() error {
	return db.Updates(u).Error
}

// 构造函数
func Make(uid string) *User {
	u := User{
		Uid:        uid,
		Token:      uuid.NewV4().String(),
		Permission: 5.10,
	}
	db.Create(&u)
	return &u
}

// 根据 uid 查询
func Query(token string) *User {
	if token == "" {
		return nil
	}
	var u User
	if !Update(&u, "token = ?", token) {
		return nil
	}
	return &u
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
	db.Model(u).Update("permission", u.Permission)
}

func (u *User) Scan(val any) error {
	return db.First(u, "uid = ?", val).Error
}

func (u User) Value() (driver.Value, error) {
	return u.Uid, nil
}
