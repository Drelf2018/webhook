package network

import (
	"strings"

	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
)

// 替换通配符
func ReplaceData(text string, post *data.Post, content string) string {
	return strings.NewReplacer(
		"{mid}", post.Mid,
		"{time}", post.Time,
		"{text}", post.Text,
		"{source}", post.Source,
		"{platform}", post.Platform,
		"{uid}", post.Uid,
		"{name}", post.Name,
		"{face}", post.Face.ToURL(),
		"{pendant}", post.Pendant.ToURL(),
		"{description}", post.Description,
		"{follower}", post.Follower,
		"{following}", post.Following,
		"{attachments}", post.Attachments.ToURL(),
		"{content}", content,
		"{repost.", "{",
	).Replace(text)
}

// 回传博文
func Webhook(post *data.Post) {
	// 获取纯净文本
	content := utils.Clean(post.Text)
	if post.Repost == nil {
		post.Repost = &data.Post{}
	}
	rcontent := utils.Clean(post.Repost.Text)

	// 正则匹配任务
	loop := utils.EventLoop[*request.Result]{}
	for _, job := range user.GetJobsByRegexp(post.Platform, post.Uid) {
		for k, v := range job.Data {
			v = ReplaceData(v, post, content)
			v = ReplaceData(v, post.Repost, rcontent)
			job.Data[k] = v
		}
		loop.AddFunc(job.Request)
	}
	loop.Wait()
}
