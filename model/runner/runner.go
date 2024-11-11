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
	TaskDB  *gorm.DB
	Timeout time.Duration

	mu sync.Mutex
}

func (t *TaskRunner) SendBlogWithContext(ctx context.Context, blog *model.Blog) error {
	t.mu.Lock()

	matched := t.TaskDB.Where(fmt.Sprintf(model.FilterQuery, blog.Submitter, blog.UID, blog.Type, blog.Platform))
	if !blog.Edited {
		matched = matched.Not(model.ExceptQuery, blog.MID, blog.Type, blog.Platform)
	}

	tasks := make([]*model.Task, 0)
	err := matched.Find(&tasks).Error
	if err != nil || len(tasks) == 0 {
		t.mu.Unlock()
		return err
	}

	logs := make([]model.RequestLog, len(tasks))
	for idx, task := range tasks {
		logs[idx] = model.RequestLog{
			TaskID:   task.ID,
			BlogID:   blog.ID,
			MID:      blog.MID,
			Type:     blog.Type,
			Platform: blog.Platform,
		}
	}

	err = t.TaskDB.Create(logs).Error
	t.mu.Unlock()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	tmpl := model.NewTemplate(blog)
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for idx := range tasks {
		go func(idx int) {
			r, err := tasks[idx].Api.DoWithContext(ctx, tmpl)
			logs[idx].Result = string(r)
			if err != nil {
				logs[idx].Error = err.Error()
			}
			wg.Done()
		}(idx)
	}
	wg.Wait()
	cancel()
	return t.TaskDB.Save(logs).Error
}

func (t *TaskRunner) SendBlog(blog *model.Blog) error {
	return t.SendBlogWithContext(context.Background(), blog)
}
