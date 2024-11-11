package api

import (
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/model/runner"

	"github.com/gin-gonic/gin"
)

// 获取自身信息
func Get(ctx *gin.Context) (any, error) {
	user := &model.User{UID: GetUID(ctx)}
	err := UserDB().Preload("Tasks").First(user).Error
	if err != nil {
		return 1, err
	}
	return user, nil
}

func GetTaskID(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	tx := UserDB().Preload("RequestLogs").Limit(1).Find(task, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	return task, nil
}

// 新增任务
func PostTask(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	err := ctx.ShouldBindJSON(task)
	if err != nil {
		return 1, err
	}
	task.UserID = GetUID(ctx)
	err = UserDB().Create(task).Error
	if err != nil {
		return 2, err
	}
	return task.ID, nil
}

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
	go func() {
		sendErr := Runner().SendBlog(blog)
		if sendErr != nil {
			Log().Errorf("%s: %s", blog, sendErr)
		}
	}()
	return blog.ID, nil
}

// 测试单个任务
func PostTest(ctx *gin.Context) (any, error) {
	var data struct {
		Blog model.Blog `json:"blog"`
		Task model.Task `json:"task"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		return 1, err
	}
	return runner.TestTaskWithContext(ctx, &data.Blog, data.Task), nil
}

// 测试已有任务
func PostTests(ctx *gin.Context) (any, error) {
	var data struct {
		Blog  model.Blog `json:"blog"`
		Tasks []uint64   `json:"tasks"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		return 1, err
	}
	var tasks []model.Task
	err = UserDB().Find(&tasks, "user_id = ? AND id in ?", GetUID(ctx), data.Tasks).Error
	if err != nil {
		return 2, err
	}
	return runner.TestTasksWithContext(ctx, &data.Blog, tasks), nil
}

// 移除任务
func DeleteTaskID(ctx *gin.Context) (any, error) {
	err := UserDB().Delete(&model.Task{}, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx)).Error
	if err != nil {
		return 1, err
	}
	return "success", nil
}
