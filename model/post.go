package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/utils"
	"github.com/itchyny/timefmt-go"
	"gorm.io/gorm"
)

var _ = os.MkdirAll("./public", os.ModePerm)
var postDB = gorms.SetSQLite("./public/post.db").AutoMigrate(&Tag{}, &Post{})
var ErrSubmitted = errors.New("您已提交过")

func Close() (error, error) {
	return userDB.Close(), postDB.Close()
}

// 博文
type Post struct {
	gorms.Model[uint64] `form:"-" json:"-"`
	// 平台
	Platform string `form:"platform" json:"platform"`
	// 平台序号
	Uid string `form:"uid" json:"uid" cmps:"1"`
	// 昵称
	Name string `form:"name" json:"name"`
	// 粉丝数
	Follower string `form:"follower" json:"follower"`
	// 关注数
	Following string `form:"following" json:"following"`
	// 简介
	Description string `form:"description" json:"description"`
	// 头像
	AvatarUrl string     `gorm:"column:avatar"`
	Avatar    Attachment `form:"avatar" json:"avatar"`
	// 装扮
	PendantUrl string     `gorm:"column:pendant"`
	Pendant    Attachment `form:"pendant" json:"pendant"`
	// 头图
	BannerUrl string     `gorm:"column:banner"`
	Banner    Attachment `form:"banner" json:"banner"`
	// 博文序号
	Mid string `form:"mid" json:"mid"`
	// 发送时间
	Time string `form:"time" json:"time"`
	// 文本
	Text string `form:"text" json:"text"`
	// 来源
	Source string `form:"source" json:"source"`
	// 附件
	Attachments []Attachment `gorm:"many2many:post_attachments" form:"attachments" json:"attachments"`
	// 标签
	Tags []Tag
	// 提交者
	Submitter *User `form:"-" json:"submitter"`
	// 回复
	RepostID *uint64 `form:"-" json:"-" gorm:"column:repost"`
	Repost   *Post   `form:"repost" json:"repost"`
	// 评论
	CommentID *uint64 `form:"-" json:"-" gorm:"column:comment"`
	Comments  []Post  `form:"comments" json:"comments" gorm:"foreignKey:CommentID"`
	// 替换器
	replacer *strings.Replacer `gorm:"-" form:"-" json:"-"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.Repost == nil {
		return nil
	}
	if reflect.DeepEqual(p.Repost, &Post{}) {
		p.Repost = nil
		return nil
	}
	p.Repost.Submitter = p.Submitter
	return nil
}

func (p *Post) Key() string {
	return p.Platform + p.Mid
}

func (p *Post) Submit() error {
	v, _ := monitors.LoadOrStore(p.Key(), &Monitor{Posts: make([]*OrderedPost, 0)})
	m := v.(*Monitor)
	if m.IsSubmitted(p.Submitter.Uid) {
		return ErrSubmitted
	}
	go m.Parse(&OrderedPost{Post: p})
	return nil
}

const maxTextLength = 18

func (p *Post) String() string {
	text := utils.Clean(p.Text)
	if len(text) > maxTextLength {
		text = text[:maxTextLength] + "..."
	}
	if p.Repost == nil {
		return fmt.Sprintf("Post(%s, %s)", p.Name, text)
	}
	return fmt.Sprintf("Post(%s, %s, %v)", p.Name, text, p.Repost)
}

// 替换通配符
func (p *Post) replaceData(text string) (s string) {
	if p.replacer == nil {
		post, _ := json.Marshal(p)
		urls, _ := json.Marshal(p.Attachments)
		p.replacer = strings.NewReplacer(
			"{mid}", p.Mid,
			"{time}", p.Time,
			"{text}", p.Text,
			"{source}", p.Source,
			"{platform}", p.Platform,
			"{uid}", p.Uid,
			"{name}", p.Name,
			"{face}", p.Avatar.Url,
			"{pendant}", p.Pendant.Url,
			"{description}", p.Description,
			"{follower}", p.Follower,
			"{following}", p.Following,
			"{attachments}", string(urls),
			"{content}", utils.Clean(p.Text),
			"{post}", string(post),
			"{repost.", "{",
		)
	}
	s = p.replacer.Replace(p.replaceTimeFormat(text))
	if p.Repost != nil {
		return p.Repost.replaceData(s)
	}
	return
}

var timeFormatter = regexp.MustCompile(`\{format:([^\}]+)\}`)

func (p *Post) replaceTimeFormat(s string) string {
	oldnew := make([]string, 0)
	tt := utils.Time{String: p.Time}.ToDate()
	for _, s := range timeFormatter.FindAllStringSubmatch(s, -1) {
		oldnew = append(oldnew, s[0], timefmt.Format(tt, s[1]))
	}
	return strings.NewReplacer(oldnew...).Replace(s)
}

func (p *Post) Send(jobs []Job) []*request.JobResult {
	return asyncio.ForEachV(jobs, func(job Job) *request.JobResult {
		for k, v := range job.Data {
			job.Data[k] = p.replaceData(v)
		}
		return job.Test()
	})
}

// 回调博文
func (p *Post) Webhook() {
	p.Send(GetJobsByRegexp(p.Platform + p.Uid))
}

// 判断博文是否存在
func (p *Post) Exists() bool {
	return gorms.Exists[Post]("platform = ? AND mid = ?", p.Platform, p.ID)
}

// // 通过平台和序号获取唯一博文
// func GetPost(platform, mid string) *Post {
// 	var p Post
// 	Posts.Preload(&p, "platform = ? AND mid = ?", platform, mid)
// 	return &p
// }

// // 获取分支
// func GetBranches(platform, mid string, r *[]Post) {
// 	Posts.Preloads(r, "platform = ? AND mid = ?", platform, mid)
// }

// 获取起始与结束时间范围内所有博文
func GetPosts(begin, end string) []Post {
	return gorms.MustPreloads[Post]("time BETWEEN ? AND ?", begin, end)
}

// 测试博文
var TestPost = &Post{
	Platform:    "weibo",
	Uid:         "7198559139",
	Name:        "七海Nana7mi",
	Avatar:      Attachment{Url: "https://wx4.sinaimg.cn/orj480/007Raq4zly8hd1vqpx3coj30u00u00uv.jpg"},
	Follower:    "104.8万",
	Following:   "192",
	Description: "蓝色饭团",
	Mid:         "4952487292307646",
	Time:        "1696248446",
	Text:        `<span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span>`,
	Source:      "🦈iPhone 14 Pro Max",
	Submitter:   &User{Uid: "188888131", Permission: Owner},
	Repost: &Post{
		Platform:    "weibo",
		Uid:         "2203177060",
		Name:        "阿梓从小就很可爱",
		Avatar:      Attachment{Url: "https://wx4.sinaimg.cn/orj480/8351d064ly8hiph621dryj20u00u00vw.jpg"},
		Follower:    "61.9万",
		Following:   "306",
		Description: "本人只喜欢读书",
		Mid:         "4952449691946355",
		Time:        "1696239481",
		Text:        "[看书] ",
		Source:      "iPhone 13 Pro Max",
		Attachments: []Attachment{{Url: "https://wx2.sinaimg.cn/large/8351d064ly1hih23pb486j21tk19k7wk.jpg"}},
		Submitter:   &User{Uid: "188888131"},
		Repost:      &Post{},
	},
}
