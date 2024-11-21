package api

import (
	"fmt"

	"github.com/Drelf2018/webhook/model"

	"github.com/gin-gonic/gin"
)

// 提交博文
func PostBlog(ctx *gin.Context) (any, error) {
	blog := &model.Blog{}
	err := ctx.ShouldBindJSON(blog)
	if err != nil {
		return 1, err
	}
	blog.Submitter = GetUID(ctx)
	err = BlogDB().Create(blog).Error
	if err != nil {
		return 2, err
	}
	go hook(ctx.Copy(), blog)
	return blog.ID, nil
}

func hook(ctx *gin.Context, blog *model.Blog) {
	var tasks []*model.Task
	err := UserDB().Find(&tasks, "enable AND id IN (?)",
		UserDB().Model(&model.Filter{}).Distinct("task_id").Where(
			`(submitter = "" OR submitter = ?) AND
			(platform = "" OR platform = ?) AND
			(type = "" OR type = ?) AND
			(uid = "" OR uid = ?)`,
			blog.Submitter,
			blog.Platform,
			blog.Type,
			blog.UID,
		),
	).Error
	if err != nil {
		Error(ctx, fmt.Errorf("webhook/api: %s: %v", blog, err))
		return
	}
	if len(tasks) == 0 {
		return
	}
	err = UserDB().Create(model.NewTemplate(blog).RunTasks(ctx, tasks)).Error
	if err != nil {
		Error(ctx, fmt.Errorf("webhook/api: %s: %v", blog, err))
	}
}

// 新增任务
func PostTask(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	err := ctx.ShouldBindJSON(task)
	if err != nil {
		return 1, err
	}
	if len(task.Filters) == 0 {
		return 2, ErrFilterNotExist
	}
	task.UserID = GetUID(ctx)
	err = UserDB().Create(task).Error
	if err != nil {
		return 3, err
	}
	return task.ID, nil
}

// 获取任务
func GetTaskID(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	tx := UserDB().Preload("Filters").Preload("Logs").Limit(1).Find(task, "id = ? AND (public OR user_id = ?)", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	return task, nil
}

// 移除任务
func DeleteTaskID(ctx *gin.Context) (any, error) {
	tx := UserDB().Delete(&model.Task{}, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	return "success", nil
}

// 获取自身信息
func Get(ctx *gin.Context) (any, error) {
	user := &model.User{UID: GetUID(ctx)}
	err := UserDB().Preload("Tasks").First(user).Error
	if err != nil {
		return 1, err
	}
	return user, nil
}

// 测试单个任务
func PostTest(ctx *gin.Context) (any, error) {
	var data struct {
		Blog *model.Blog `json:"blog"`
		Task *model.Task `json:"task"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		return 1, err
	}
	if data.Blog == nil {
		return 2, ErrBlogNotExist
	}
	if data.Task == nil {
		return 3, ErrTaskNotExist
	}
	return model.NewTemplate(data.Blog).RunTask(ctx, data.Task), nil
}

// 测试已有任务
func PostTests(ctx *gin.Context) (any, error) {
	var data struct {
		Blog  *model.Blog `json:"blog"`
		Tasks []uint64    `json:"tasks"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		return 1, err
	}
	if data.Blog == nil {
		return 2, ErrBlogNotExist
	}
	var tasks []*model.Task
	err = UserDB().Find(&tasks, "user_id = ? AND id in ?", GetUID(ctx), data.Tasks).Error
	if err != nil {
		return 3, err
	}
	return model.NewTemplate(data.Blog).RunTasks(ctx, tasks), nil
}
