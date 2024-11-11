package runner

import (
	"context"
	"sync"
	"time"

	"github.com/Drelf2018/webhook/model"
)

func TestTaskWithContext(ctx context.Context, blog *model.Blog, task model.Task) model.RequestLog {
	r, err := task.Api.DoWithContext(ctx, model.NewTemplate(blog))
	log := model.RequestLog{
		CreatedAt: time.Now(),
		TaskID:    task.ID,
		BlogID:    blog.ID,
		MID:       blog.MID,
		Type:      blog.Type,
		Platform:  blog.Platform,
		Result:    string(r),
	}
	if err != nil {
		log.Error = err.Error()
	}
	return log
}

func TestTasksWithContext(ctx context.Context, blog *model.Blog, tasks []model.Task) []model.RequestLog {
	logs := make([]model.RequestLog, len(tasks))
	tmpl := model.NewTemplate(blog)
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for idx := range tasks {
		go func(idx int) {
			r, err := tasks[idx].Api.DoWithContext(ctx, tmpl)
			logs[idx] = model.RequestLog{
				CreatedAt: time.Now(),
				TaskID:    tasks[idx].ID,
				BlogID:    blog.ID,
				MID:       blog.MID,
				Type:      blog.Type,
				Platform:  blog.Platform,
				Result:    string(r),
			}
			if err != nil {
				logs[idx].Error = err.Error()
			}
			wg.Done()
		}(idx)
	}
	wg.Wait()
	return logs
}
