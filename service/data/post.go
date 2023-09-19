package data

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	"gorm.io/gorm"
)

// 博文或评论
type Post struct {
	Model
	// 平台
	Platform string `form:"platform" json:"platform"`
	// 博文序号
	Mid string `form:"mid" json:"mid"`
	// 发送时间
	Time string `form:"time" json:"time"`
	// 文本
	Text string `form:"text" json:"text"`
	// 内容
	Content string `gorm:"-" form:"-" json:"-"`
	// 来源
	Source string `form:"source" json:"source"`
	// 博主
	BloggerID string  `form:"-" json:"-" gorm:"column:blogger"`
	Blogger   Blogger `form:"blogger" json:"blogger" preload:""`
	// 回复
	RepostID *uint64 `form:"-" json:"-" gorm:"column:repost"`
	Repost   *Post   `form:"repost" json:"repost"`
	// 评论
	CommentID *uint64 `form:"-" json:"-" gorm:"column:comment"`
	Comments  []Post  `form:"comments" json:"comments" gorm:"foreignKey:CommentID"`
	// 附件
	Attachments Attachments `form:"attachments" json:"attachments" gorm:"many2many:post_attachments;"`
	// 提交者
	Submitter *user.User `form:"-" json:"submitter"`
	// 编辑距离
	Distance int `gorm:"-" form:"-" json:"-" cmps:"1;order:desc"`
	// 替换器
	Replacer *strings.Replacer `gorm:"-" form:"-" json:"-"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if reflect.DeepEqual(p.Repost, &Post{}) {
		p.Repost = nil
	} else {
		p.Repost.Submitter = p.Submitter
	}
	asyncio.ForEachP(p.Comments, func(c *Post) {
		c.Submitter = p.Submitter
	})
	return nil
}

func (p *Post) Type() string {
	return p.Platform + p.Mid
}

func (p Post) String() string {
	return fmt.Sprintf(
		"Post(id=%v, platform=%v, text=%v, blogger=%v, repost=%v, comments=%v, attachments=%v)",
		p.ID, p.Platform, p.Text, p.Blogger, p.Repost, p.Comments, p.Attachments,
	)
}

// 替换通配符
func (p *Post) ReplaceData(text string) string {
	if p.Replacer == nil {
		p.Replacer = strings.NewReplacer(
			"{mid}", p.Mid,
			"{time}", p.Time,
			"{text}", p.Text,
			"{source}", p.Source,
			"{platform}", p.Blogger.Platform,
			"{uid}", p.Blogger.Uid,
			"{name}", p.Blogger.Name,
			"{face}", p.Blogger.Face.Url,
			"{pendant}", p.Blogger.Pendant.Url,
			"{description}", p.Blogger.Description,
			"{follower}", p.Blogger.Follower,
			"{following}", p.Blogger.Following,
			"{attachments}", p.Attachments.Urls(),
			"{content}", p.Content,
			"{repost.", "{",
		)
	}
	return p.Replacer.Replace(text)
}

// 回调博文
func (p *Post) Webhook() {
	jobs := user.GetJobsByRegexp(p.Platform, p.Blogger.Uid)
	asyncio.ForEach(jobs, func(job user.Job) {
		for k, v := range job.Data {
			v = p.ReplaceData(v)
			if p.Repost != nil {
				v = p.Repost.ReplaceData(v)
			}
			job.Data[k] = v
		}
		asyncio.RetryError(3, 5, func() error { return job.Request().Error })
	})
}

func SavePost(p *Post) {
	p.Content = utils.Clean(p.Text)
	p.Submitter.LevelUP()
	db.Create(p)
	go p.Webhook()
}

func SavePosts(p ...*Post) {
	if len(p) == 0 {
		return
	}
	asyncio.ForEach(p, func(p *Post) { p.Submitter.LevelUP() })
	db.Create(&p)
}

// 判断博文是否存在
func HasPost(platform, mid string) bool {
	return Exists[Post]("platform = ? AND mid = ?", platform, mid)
}

// 通过平台和序号获取唯一博文
func GetPost(platform, mid string) *Post {
	var p Post
	Preload(&p, "platform = ? AND mid = ?", platform, mid)
	return &p
}

// 获取分支
func GetBranches(platform, mid string, r *[]Post) {
	Preloads(r, "platform = ? AND mid = ?", platform, mid)
}

// 通过起始与结束时间获取范围内博文合集
func GetPosts(begin, end string, r *[]Post) {
	Preloads(r, "time BETWEEN ? AND ?", begin, end)
}
