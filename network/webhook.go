package network

import (
	"strings"

	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
)

// 替换通配符
func ReplaceData(text string, post *data.Post) string {
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
		"{content}", post.Content,
		"{repost.", "{",
	).Replace(text)
}

// 回调博文
func Webhook(p *data.Post) {
	jobs := user.GetJobsByRegexp(p.Platform, p.Uid)
	utils.Await(
		func(job user.Job) *request.Result {
			for k, v := range job.Data {
				v = ReplaceData(v, p)
				if p.Repost != nil {
					v = ReplaceData(v, p.Repost)
				}
				job.Data[k] = v
			}
			return job.Request()
		},
		&jobs,
	)
}
