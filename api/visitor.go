package api

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// 注册账户
func PostRegister(ctx *gin.Context) (any, error) {
	u, data, err := registrar.Register(ctx)
	if u == nil {
		return data, err
	}
	tx := UserDB.Limit(1).Find(u)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected != 0 {
		return 2, ErrUserRegistered
	}
	// 新建用户
	user := u.(*model.User)
	if user.UID == config.Role.Owner {
		user.Role = model.Owner
	} else {
		user.Role = model.Normal
		for _, admin := range config.Role.Admin {
			if user.UID == admin {
				user.Role = model.Admin
				break
			}
		}
	}
	err = UserDB.Create(user).Error
	if err != nil {
		return 3, err
	}
	return Success, nil
}

// 获取 Token
func GetToken(ctx *gin.Context) (data any, err error) {
	uid, password, err := registrar.BasicAuth(ctx)
	if err != nil {
		return 1, err
	}
	user := &model.User{UID: uid}
	tx := UserDB.Limit(1).Find(user)
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
	var iat any
	var found bool
	claim := UserClaims{uid, now.UnixMilli()}
	if ctx.Query("refresh") != "true" {
		// 获取已有的 Token
		iat, found = tokenIssuedAt.Load(uid)
		if found {
			claim.IssuedAt = iat.(int64)
		}
	}
	token, err := claim.Token(!found)
	if err != nil {
		return 6, err
	}
	return token, nil
}

// 获取用户信息
func GetUserUID(ctx *gin.Context) (any, error) {
	user := &model.User{UID: ctx.Param("uid")}
	uid, _ := JWTAuth(ctx)
	tx := UserDB.Limit(1)
	if uid == user.UID || uid == config.Role.Owner {
		tx = tx.Preload("Tasks").Preload("Tasks.Filters")
	}
	tx = tx.Find(user)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrUserNotExist
	}
	return user, nil
}

// 查询条件
type Condition struct {
	// 筛选
	Filters []model.Filter `json:"filters" form:"-"`

	// 是否包含转发
	Reply bool `json:"reply" form:"reply"`

	// 是否包含评论
	Comments bool `json:"comments" form:"comments"`

	// 查询排列顺序
	Order string `json:"order" form:"order"`

	// 查询行数
	Limit int `json:"limit" form:"limit"`

	// 查询偏移
	Offset int `json:"offset" form:"offset"`

	// 其他条件
	Conds []string `json:"conds" form:"conds"`
}

// 条件查询博文
func (c *Condition) Find(tx *gorm.DB, dest any) error {
	if c.Reply {
		tx = tx.Preload("Reply")
	}
	if c.Comments {
		tx = tx.Preload("Comments")
	}
	if c.Order != "" {
		tx = tx.Order(c.Order)
	}
	switch {
	case c.Limit > 1000:
		tx = tx.Limit(1000)
	case c.Limit > 0:
		tx = tx.Limit(c.Limit)
	default:
		tx = tx.Limit(30)
	}
	if c.Offset != 0 {
		tx = tx.Offset(c.Offset)
	}
	filter := BlogDB.Model(&model.Blog{})
	for _, f := range c.Filters {
		f.TaskID = 0
		filter = filter.Or(f)
	}
	return tx.Where(filter).Find(dest, utils.StrToAny(c.Conds)...).Error
}

func (c *Condition) Finds(tx *gorm.DB) (blogs []model.Blog, err error) {
	err = c.Find(tx, &blogs)
	return
}

// 筛选查询
func PostFilters(ctx *gin.Context) (any, error) {
	c := Condition{
		Reply:    true,
		Comments: true,
		Order:    "time desc",
	}
	err := ctx.ShouldBindJSON(&c)
	if err != nil {
		return 1, err
	}
	blogs, err := c.Finds(BlogDB)
	if err != nil {
		return 2, err
	}
	return blogs, nil
}

// 任务驱动查询
func GetTasks(ctx *gin.Context) (any, error) {
	c := struct {
		Condition
		ID []uint64 `form:"id"`
	}{
		Condition: Condition{
			Reply:    true,
			Comments: true,
			Order:    "time desc",
		},
	}
	err := ctx.ShouldBindQuery(&c)
	if err != nil {
		return 1, err
	}

	uid, _ := JWTAuth(ctx)
	taskID := UserDB.Model(&model.Task{}).Distinct("id").Where("(public OR user_id = ?) AND id IN ?", uid, c.ID)
	err = UserDB.Find(&c.Filters, "task_id IN (?)", taskID).Error
	if err != nil {
		return 2, err
	}

	blogs, err := c.Finds(BlogDB)
	if err != nil {
		return 3, err
	}
	return blogs, nil
}

// 查询博文
func GetBlogs(ctx *gin.Context) (any, error) {
	q := struct {
		Condition
		model.Filter
		MID string `form:"mid" gorm:"column:mid"`
	}{
		Condition: Condition{
			Reply:    true,
			Comments: true,
			Order:    "time desc",
		},
	}
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		return 1, err
	}
	q.Condition.Filters = []model.Filter{q.Filter}
	blogs, err := q.Finds(BlogDB)
	if err != nil {
		return 2, err
	}
	return blogs, nil
}

// 查询单条博文
func GetBlogID(ctx *gin.Context) (any, error) {
	blog := &model.Blog{}
	tx := BlogDB.Preload("Reply").Preload("Comments").Limit(1).Find(blog, "id = ?", ctx.Param("id"))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrBlogNotExist
	}
	return blog, nil
}
