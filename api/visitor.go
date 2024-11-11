package api

import (
	"errors"
	"sync"
	"time"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"

	_ "unsafe"
)

const Version = "v0.15.0"

var (
	ErrUserRegistered = errors.New("webhook/api: user registered")
	ErrUserNotExist   = errors.New("webhook/api: user does not exist")
	ErrIncorrectPwd   = errors.New("webhook/api: incorrect password")
	ErrBanned         = errors.New("webhook/api: user has been banned")
	ErrBlogNotExist   = errors.New("webhook/api: blog does not exist")
	ErrTaskNotExist   = errors.New("webhook/api: task does not exist")
)

var onlineUsers sync.Map //map[string]time.Time

// 当前版本号
func GetVersion(ctx *gin.Context) (any, error) {
	return Version, nil
}

// 更新在线时间
func GetPing(ctx *gin.Context) (any, error) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		Error(ctx, err)
		return 1, err
	}
	onlineUsers.Store(uid, time.Now())
	return "pong", nil
}

// 获取当前在线状态
func GetOnline(ctx *gin.Context) (any, error) {
	now := time.Now()
	m := make(map[string]int64)
	onlineUsers.Range(func(key, value any) bool {
		m[key.(string)] = now.Sub(value.(time.Time)).Milliseconds()
		return true
	})
	return m, nil
}

// 获取 Token
func GetToken(ctx *gin.Context) (data any, err error) {
	uid, password, err := registrar.BasicAuth(ctx)
	if err != nil {
		return 1, err
	}
	user := &model.User{UID: uid}
	tx := UserDB().Limit(1).Find(user)
	if tx.Error != nil {
		return 2, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 3, ErrUserNotExist
	}
	if user.Password != password {
		return 4, ErrIncorrectPwd
	}
	now := time.Now()
	if user.Ban.After(now) {
		return 5, ErrBanned
	}
	token, err := UserClaims{uid, now.UnixMilli()}.Token()
	if err != nil {
		return 6, err
	}
	return token, nil
}

// 获取用户信息
func GetUUID(ctx *gin.Context) (any, error) {
	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB().Limit(1).Find(user)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrUserNotExist
	}
	return user, nil
}

type BlogQuery struct {
	ID        uint64   `form:"id"`
	Submitter string   `form:"submitter"`
	Platform  string   `form:"platform"`
	Type      string   `form:"type"`
	UID       string   `form:"uid"`
	MID       string   `form:"mid" gorm:"column:mid"`
	Reply     bool     `form:"reply" gorm:"-"`
	Comments  bool     `form:"comments" gorm:"-"`
	Offset    int      `form:"offset" gorm:"-"`
	Limit     int      `form:"limit" gorm:"-"`
	Conds     []string `form:"conds" gorm:"-"`
}

// 查询博文
func GetBlogs(ctx *gin.Context) (any, error) {
	var q BlogQuery
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		return 1, err
	}
	tx := BlogDB()
	if q.Reply {
		tx = tx.Preload("Reply")
	}
	if q.Comments {
		tx = tx.Preload("Comments")
	}
	tx = tx.Where(q)
	if q.Offset != 0 {
		tx = tx.Offset(q.Offset)
	}
	switch {
	case q.Limit >= 100:
		tx = tx.Limit(100)
	case q.Limit > 0:
		tx = tx.Limit(q.Limit)
	default:
		tx = tx.Limit(30)
	}
	var blogs []model.Blog
	err = tx.Find(&blogs, utils.StrToAny(q.Conds)...).Error
	if err != nil {
		return 2, err
	}
	return blogs, nil
}

// 查询单条博文
func GetBlogID(ctx *gin.Context) (any, error) {
	blog := &model.Blog{}
	tx := BlogDB().Preload("Reply").Preload("Comments").Limit(1).Find(blog, "id = ?", ctx.Param("id"))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrBlogNotExist
	}
	return blog, nil
}

// 新建用户
func PostRegister(ctx *gin.Context) (any, error) {
	u, data, err := registrar.Register(ctx)
	if u == nil {
		return data, err
	}
	tx := UserDB().Limit(1).Find(u)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected != 0 {
		return 2, ErrUserRegistered
	}
	// create user
	user := u.(*model.User)
	if user.UID == webhook.Global().Role.Owner {
		user.Role = model.Owner
	} else {
		user.Role = model.Normal
		for _, admin := range webhook.Global().Role.Admin {
			if user.UID == admin {
				user.Role = model.Admin
				break
			}
		}
	}
	err = UserDB().Create(user).Error
	if err != nil {
		return 3, err
	}
	return "success", nil
}
