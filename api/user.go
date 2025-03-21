package api

import (
	"errors"
	"fmt"

	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/utils"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// 下载资源
func DownloadAssets(blog *model.Blog) error {
	if blog == nil {
		return nil
	}
	errs := make(utils.JoinError, 0)
	if blog.Avatar != "" {
		_, err := downloader.Download(blog.Avatar)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, url := range blog.Assets {
		_, err := downloader.Download(url)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, url := range blog.Banner {
		_, err := downloader.Download(url)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if blog.Reply != nil {
		err := DownloadAssets(blog.Reply)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// 关注中博文查询
func GetFollowing(ctx *gin.Context) (any, error) {
	f := Filter{
		Reply:    true,
		Comments: true,
		Order:    "time desc",
	}
	err := ctx.ShouldBindQuery(&f)
	if err != nil {
		return 1, err
	}
	err = UserDB.Where("task_id in (?)", UserDB.Model(&model.Task{}).Distinct("id").Where("user_id = ?", GetUID(ctx))).Find(&f.Filters).Error
	if err != nil {
		return 2, err
	}
	var blogs []model.Blog
	err = FindBlogs(f, &blogs)
	if err != nil {
		return 3, err
	}
	return blogs, nil
}

// 提交博文
func PostBlog(ctx *gin.Context) (any, error) {
	blog := &model.Blog{}
	err := ctx.ShouldBindJSON(blog)
	if err != nil {
		return 1, err
	}
	blog.Submitter = GetUID(ctx)
	err = BlogDB.Create(blog).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 2, ErrBlogNotExist
	}
	if err != nil {
		return 3, err
	}
	var tasks []*model.Task
	err = UserDB.Find(&tasks, "enable AND id IN (?)",
		UserDB.Model(&model.Filter{}).Distinct("task_id").Where(
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
		return 4, fmt.Errorf("webhook/api: %s: %v", blog, err)
	}
	go func(c *gin.Context) {
		if len(tasks) != 0 {
			err := UserDB.Create(model.NewTemplate(blog).RunTasks(tasks)).Error
			if err != nil {
				Error(c, err)
			}
		}
		if AutoDownload {
			err := DownloadAssets(blog)
			if err != nil {
				Error(c, err)
			}
		}
	}(ctx.Copy())
	return blog.ID, nil
}

func removeDuplicatesFilters(filters []model.Filter) (result []model.Filter) {
	exists := make(map[string]struct{}, len(filters))
	for _, f := range filters {
		if _, ok := exists[f.String()]; ok {
			continue
		}
		exists[f.String()] = struct{}{}
		if !f.IsZero() {
			result = append(result, f)
		}
	}
	return
}

// 新增任务
func PostTask(ctx *gin.Context) (any, error) {
	task := &model.Task{}
	err := ctx.ShouldBindJSON(task)
	if err != nil {
		return 1, err
	}
	task.Filters = removeDuplicatesFilters(task.Filters)
	if len(task.Filters) == 0 {
		return 2, ErrFilterNotExist
	}
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
	tx := UserDB.Preload("Filters").Preload("Logs").Limit(1).Find(task, "id = ? AND (public OR user_id = ?)", ctx.Param("id"), GetUID(ctx))
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
	tx := UserDB.Delete(&model.Task{}, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
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
	err := UserDB.Preload("Tasks").Preload("Tasks.Filters").First(user).Error
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
	return model.NewTemplate(data.Blog).RunTask(data.Task), nil
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
	err = UserDB.Find(&tasks, "user_id = ? AND id in ?", GetUID(ctx), data.Tasks).Error
	if err != nil {
		return 3, err
	}
	return model.NewTemplate(data.Blog).RunTasks(tasks), nil
}
