package api

import (
	"errors"
	"fmt"
	"net"

	"github.com/Drelf2018/webhook/model"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// 获取自身信息
func GetUser(ctx *gin.Context) (any, error) {
	user := &model.User{UID: GetUID(ctx)}
	err := UserDB.Preload("Tasks").Preload("Tasks.Filters").First(user).Error
	if err != nil {
		return 1, err
	}
	return user, nil
}

// 获取关注的博文
func GetFollowing(ctx *gin.Context) (any, error) {
	c := &Condition{
		Reply:    true,
		Comments: true,
		Order:    "time desc",
	}
	err := ctx.ShouldBindQuery(c)
	if err != nil {
		return 1, err
	}
	var filters []model.Filter
	taskID := UserDB.Model(&model.Task{}).Distinct("id").Where("user_id = ?", GetUID(ctx))
	err = UserDB.Find(&filters, "task_id IN (?)", taskID).Error
	if err != nil {
		return 2, err
	}
	blogs, err := c.Finds(BlogDB, filters...)
	if err != nil {
		return 3, err
	}
	return blogs, nil
}

// 提交博文
func PostBlog(ctx *gin.Context) (any, error) {
	// 避免回环提交
	ip := net.ParseIP(ctx.ClientIP())
	if ip == nil || ip.IsLoopback() {
		return 1, fmt.Errorf("webhook/api: client IP error: %v", ip)
	}

	// 绑定博文并写入提交者
	// 防止有人打着别人名号提交
	blog := &model.Blog{}
	err := ctx.ShouldBindJSON(blog)
	if err != nil {
		return 2, err
	}
	blog.Submitter = GetUID(ctx)

	// (*Blog).AfterCreate 在没找到它回复的博文时会返回 gorm.ErrRecordNotFound 错误
	err = BlogDB.Create(blog).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 3, ErrBlogNotExist
	} else if err != nil {
		return 4, err
	}

	// 似乎在异步中操作数据库的话耗时会激增
	// 所以只能苦一苦提交者
	// 骂名阿里云来背
	var tasks []*model.Task
	err = UserDB.Raw(TasksQuery, blog.Submitter, blog.Platform, blog.Type, blog.UID).Find(&tasks).Error
	if err != nil {
		Error(ctx, fmt.Errorf("webhook/api: query %s error: %w", blog, err))
	}

	if len(tasks) != 0 {
		err = UserDB.Create(model.NewTemplate(blog).RunTasks(tasks)).Error
		if err != nil {
			Error(ctx, fmt.Errorf("webhook/api: webhook %s error: %w", blog, err))
		}
	}

	if AutoDownload {
		err = DownloadAssets(blog)
		if err != nil {
			Error(ctx, fmt.Errorf("webhook/api: auto download error: %w", err))
		}
	}

	return blog.ID, nil
}

// 筛选条件去重
func DeduplicateFilters(filters []model.Filter) (result []model.Filter) {
	exists := make(map[model.Filter]struct{}, len(filters))
	for _, f := range filters {
		if _, ok := exists[f]; !ok {
			exists[f] = struct{}{}
			if !f.IsZero() {
				result = append(result, f)
			}
		}
	}
	return
}

// 提交任务
func PostTask(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	err := ctx.ShouldBindJSON(task)
	if err != nil {
		return 1, err
	}
	if len(task.Filters) == 0 {
		return 2, ErrFilterNotExist
	}
	task.Filters = DeduplicateFilters(task.Filters)
	task.UserID = GetUID(ctx)
	err = UserDB.Create(task).Error
	if err != nil {
		return 3, err
	}
	return task.ID, nil
}

// 获取任务
func GetTaskID(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	tx := UserDB.Preload("Filters").Limit(1).Find(task, "id = ?", ctx.Param("id"))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	uid := GetUID(ctx)
	if task.UserID != uid && !task.Public && uid != config.Role.Owner {
		return 3, ErrPermDenied
	}
	var q struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		return 4, err
	}
	switch {
	case q.Limit > 1000:
		q.Limit = 1000
	case q.Limit <= 0:
		q.Limit = 30
	}
	err = UserDB.Order("created_at desc").Offset(q.Offset).Limit(q.Limit).Find(&task.Logs, "task_id = ?", task.ID).Error
	if err != nil {
		return 5, err
	}
	return task, nil
}

// 移除任务
func DeleteTaskID(ctx *gin.Context) (any, error) {
	tx := UserDB.Delete(&model.Task{}, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	return Success, nil
}

// 测试任务
func PostTest(ctx *gin.Context) (any, error) {
	var data struct {
		Blog   *model.Blog `json:"blog"`
		Task   *model.Task `json:"task"`
		BlogID uint64      `json:"blog_id"`
		TaskID []uint64    `json:"task_id"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		return 1, err
	}
	if data.BlogID != 0 {
		tx := BlogDB.Limit(1).Find(&data.Blog, "id = ?", data.BlogID)
		if tx.Error != nil {
			return 2, tx.Error
		}
		if tx.RowsAffected == 0 {
			return 3, ErrBlogNotExist
		}
	}
	if data.Blog == nil {
		return 4, ErrBlogNotExist
	}
	// 直接测试
	if len(data.TaskID) == 0 {
		if data.Task == nil {
			return 5, ErrTaskNotExist
		}
		return []model.RequestLog{model.NewTemplate(data.Blog).RunTask(data.Task)}, nil
	}
	// 查找任务
	var tasks []*model.Task
	err = UserDB.Find(&tasks, "user_id = ? AND id IN ?", GetUID(ctx), data.TaskID).Error
	if err != nil {
		return 6, err
	}
	return model.NewTemplate(data.Blog).RunTasks(tasks), nil
}
