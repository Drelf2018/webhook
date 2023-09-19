package user

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"

	"github.com/Drelf2018/request"
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
	db.AutoMigrate(new(Job))
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

func CreateTestUser() {
	if db.First(&User{}, "uid = ?", "188888131").Error == nil {
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

// 用户
type User struct {
	Uid        string  `gorm:"primaryKey" json:"uid"`
	Token      string  `json:"-"`
	Permission float64 `json:"permission"`
	Jobs       []Job   `gorm:"references:Uid" form:"jobs" json:"-"`
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
