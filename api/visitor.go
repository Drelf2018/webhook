package api

import (
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 查询条件
type Condition struct {
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
func (c *Condition) Find(tx *gorm.DB, dest any, filters ...model.Filter) error {
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
	for _, f := range filters {
		f.TaskID = 0
		filter = filter.Or(f)
	}
	return tx.Where(filter).Find(dest, utils.StrToAny(c.Conds)...).Error
}

func (c *Condition) Finds(tx *gorm.DB, filters ...model.Filter) (blogs []model.Blog, err error) {
	err = c.Find(tx, &blogs, filters...)
	return
}

// 获取博文
func GetBlogs(ctx *gin.Context) (any, error) {
	c := &struct {
		Condition
		model.Filter
		MID    string   `form:"mid"`
		TaskID []uint64 `form:"task_id"`
	}{
		Condition: Condition{
			Reply:    true,
			Comments: true,
			Order:    "time desc",
		},
	}
	err := ctx.ShouldBindQuery(c)
	if err != nil {
		return 1, err
	}
	// 直接查询
	if len(c.TaskID) == 0 {
		tx := BlogDB
		if c.MID != "" {
			tx = tx.Where("mid = " + c.MID)
		}
		blogs, err := c.Finds(tx, c.Filter)
		if err != nil {
			return 2, err
		}
		return blogs, nil
	}
	// 从任务中获取筛选条件
	var filters []model.Filter
	// 所有者越权
	var taskID *gorm.DB
	uid, _ := JWTAuth(ctx)
	if uid == config.Role.Owner {
		taskID = UserDB.Model(&model.Task{}).Distinct("id").Where("id IN ?", c.TaskID)
	} else {
		taskID = UserDB.Model(&model.Task{}).Distinct("id").Where("(public OR user_id = ?) AND id IN ?", uid, c.TaskID)
	}
	// 合并筛选条件
	err = UserDB.Find(&filters, "task_id IN (?)", taskID).Error
	if err != nil {
		return 3, err
	}
	if len(filters) == 0 {
		return 4, ErrFilterNotExist
	}
	blogs, err := c.Finds(BlogDB, filters...)
	if err != nil {
		return 5, err
	}
	return blogs, nil
}

// 获取筛选后博文
func PostBlogs(ctx *gin.Context) (any, error) {
	c := &struct {
		Condition
		Filters []model.Filter `json:"filters"`
	}{
		Condition: Condition{
			Reply:    true,
			Comments: true,
			Order:    "time desc",
		},
	}
	err := ctx.ShouldBindJSON(c)
	if err != nil {
		return 1, err
	}
	blogs, err := c.Finds(BlogDB, c.Filters...)
	if err != nil {
		return 2, err
	}
	return blogs, nil
}

// 获取单条博文
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

// 获取任务集
func GetTasks(ctx *gin.Context) (any, error) {
	var q struct {
		Key    string `form:"key"`
		Limit  int    `form:"limit"`
		Offset int    `form:"offset"`
	}
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		return 1, err
	}
	switch {
	case q.Limit > 100:
		q.Limit = 100
	case q.Limit <= 0:
		q.Limit = 30
	}
	tx := UserDB.Preload("Filters").Order("created_at desc").Offset(q.Offset).Limit(q.Limit)
	if uid, _ := JWTAuth(ctx); uid != config.Role.Owner {
		tx = tx.Where("public")
	}
	if q.Key != "" {
		tx.Where("name LIKE ? OR readme LIKE ?", "%"+q.Key+"%", "%"+q.Key+"%")
	}
	var tasks []model.Task
	err = tx.Find(&tasks).Error
	if err != nil {
		return 2, err
	}
	return tasks, nil
}

// 注册账户
func PostUser(ctx *gin.Context) (any, error) {
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
