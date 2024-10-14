package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Drelf2018/webhook/model"
	"gorm.io/gorm"
)

type TaskRunner struct {
	DB      *gorm.DB
	Timeout time.Duration

	mu sync.Mutex
}

func (t *TaskRunner) FilterTasks(blog *model.Blog) (tasks []*model.Task, logs []model.RequestLog, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	matched := t.DB.Where(fmt.Sprintf(model.FilterQuery, blog.Submitter, blog.UID, blog.Type, blog.Platform))

	if !blog.Edited {
		matched = matched.Not(model.ExceptQuery, blog.MID, blog.Type, blog.Platform)
	}

	err = matched.Find(&tasks).Error
	if err != nil {
		return
	}

	logs = make([]model.RequestLog, len(tasks))
	for idx, task := range tasks {
		logs[idx] = model.RequestLog{
			TaskID:   task.ID,
			BlogID:   blog.ID,
			MID:      blog.MID,
			Type:     blog.Type,
			Platform: blog.Platform,
		}
	}

	err = t.DB.Create(logs).Error
	return
}

func (t *TaskRunner) SendBlogWithContext(ctx context.Context, blog *model.Blog) error {
	tasks, logs, err := t.FilterTasks(blog)
	if err != nil || len(tasks) == 0 {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	tmpl := model.NewTemplate(blog)
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for idx := range tasks {
		go func(idx int) {
			logs[idx].Result, logs[idx].Error = tasks[idx].Api.DoWithContext(ctx, tmpl)
			wg.Done()
		}(idx)
	}
	wg.Wait()
	cancel()
	return t.DB.Save(logs).Error
}

func (t *TaskRunner) SendBlog(blog *model.Blog) error {
	return t.SendBlogWithContext(context.Background(), blog)
}
