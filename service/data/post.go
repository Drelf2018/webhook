package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/service/db"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/itchyny/timefmt-go"
)

var ErrSubmitted = errors.New("您已提交过")

// 博文或评论
type Post struct {
	db.Model
	// 平台
	Platform string `form:"platform" json:"platform"`
	// 博文序号
	Mid string `form:"mid" json:"mid"`
	// 发送时间
	Time string `form:"time" json:"time"`
	// 文本
	Text string `form:"text" json:"text"`
	// 来源
	Source string `form:"source" json:"source"`
	// 博主
	BloggerID string  `form:"-" json:"-" gorm:"column:blogger"`
	Blogger   Blogger `form:"blogger" json:"blogger" default:"SetPlatform;initial.Default"`
	// 回复
	RepostID *uint64 `form:"-" json:"-" gorm:"column:repost"`
	Repost   *Post   `form:"repost" json:"repost" binding:"omitempty" default:"SetRepost;initial.Default"`
	// 评论
	CommentID *uint64 `form:"-" json:"-" gorm:"column:comment"`
	Comments  []Post  `form:"comments" json:"comments" gorm:"foreignKey:CommentID" default:"range.SetSubmitter;range.initial.Default"`
	// 附件
	Attachments Attachments `form:"attachments" json:"attachments" gorm:"many2many:post_attachments;" default:"range.Save"`
	// 提交者
	Submitter *user.User `form:"-" json:"submitter"`
	// 编辑距离
	Distance int `gorm:"-" form:"-" json:"-" cmps:"1"`
	// 替换器
	replacer *strings.Replacer `gorm:"-" form:"-" json:"-"`
}

func (p *Post) SetSubmitter(parent *Post) {
	p.Submitter = parent.Submitter
}

func (p *Post) SetRepost(parent *Post) error {
	if p == nil || reflect.DeepEqual(p, &Post{}) {
		parent.Repost = nil
		return initial.ErrBreak
	}
	p.SetSubmitter(parent)
	return nil
}

func (p *Post) Type() string {
	return p.Platform + p.Mid
}

func (p *Post) Parse() error {
	v, _ := monitors.LoadOrStore(p.Type(), &Monitor{Posts: make([]*Post, 0)})
	m := v.(*Monitor)
	if m.IsSubmitted(p.Submitter.Uid) {
		return ErrSubmitted
	}
	go m.Parse(p)
	return nil
}

func (p *Post) String() string {
	return fmt.Sprintf(
		"Post(id=%v, platform=%v, text=%v, blogger=%v, comments=%v, attachments=%v, \n  repost=%v)",
		p.ID, p.Platform, p.Text, p.Blogger, p.Comments, p.Attachments, p.Repost,
	)
}

// 替换通配符
func (p *Post) replaceData(text string) string {
	if p.replacer == nil {
		post, _ := json.Marshal(p)
		p.replacer = strings.NewReplacer(
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
			"{content}", utils.Clean(p.Text),
			"{post}", string(post),
			"{repost.", "{",
		)
	}
	return p.replacer.Replace(text)
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

func (p *Post) Send(jobs []user.Job) []*request.ResultWithJob {
	return asyncio.ForEachV(jobs, func(job user.Job) *request.ResultWithJob {
		for k, v := range job.Data {
			v = p.replaceData(p.replaceTimeFormat(v))
			if p.Repost != nil {
				v = p.Repost.replaceData(p.Repost.replaceTimeFormat(v))
			}
			job.Data[k] = v
		}
		return job.RequestWithJob()
	})
}

// 回调博文
func (p *Post) Webhook() {
	p.Send(user.GetJobsByRegexp(p.Blogger.Platform, p.Blogger.Uid))
}

// 判断博文是否存在
func HasPost(platform, mid string) bool {
	return Posts.First(new(Post), "platform = ? AND mid = ?", platform, mid)
}

// 通过平台和序号获取唯一博文
func GetPost(platform, mid string) *Post {
	var p Post
	Posts.Preload(&p, "platform = ? AND mid = ?", platform, mid)
	return &p
}

// 获取分支
func GetBranches(platform, mid string, r *[]Post) {
	Posts.Preloads(r, "platform = ? AND mid = ?", platform, mid)
}

// 通过起始与结束时间获取范围内博文合集
func GetPosts(begin, end string, r *[]Post) {
	Posts.Preloads(r, "time BETWEEN ? AND ?", begin, end)
}

// 测试博文
var TestPost = &Post{
	Platform: "weibo",
	Mid:      "4952487292307646",
	Time:     "1696248446",
	Text:     `<span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span>`,
	Source:   "🦈iPhone 14 Pro Max",
	Blogger: Blogger{
		Platform:    "weibo",
		Uid:         "7198559139",
		Name:        "七海Nana7mi",
		Create:      "1699116457",
		Follower:    "104.8万",
		Following:   "192",
		Description: "蓝色饭团",
		Face: Attachment{
			Url: "https://wx4.sinaimg.cn/orj480/007Raq4zly8hd1vqpx3coj30u00u00uv.jpg",
		},
	},
	Repost: &Post{
		Platform: "weibo",
		Mid:      "4952449691946355",
		Time:     "1696239481",
		Text:     "[看书] ",
		Source:   "iPhone 13 Pro Max",
		Blogger: Blogger{
			Platform:    "weibo",
			Uid:         "2203177060",
			Name:        "阿梓从小就很可爱",
			Create:      "1699116457",
			Follower:    "61.9万",
			Following:   "306",
			Description: "本人只喜欢读书",
			Face: Attachment{
				Url: "https://wx4.sinaimg.cn/orj480/8351d064ly8hiph621dryj20u00u00vw.jpg",
			},
		},
		Comments: make([]Post, 0),
		Attachments: Attachments{
			{Url: "https://wx2.sinaimg.cn/large/8351d064ly1hih23pb486j21tk19k7wk.jpg"},
		},
		Submitter: &user.User{Uid: "188888131", Permission: 1},
	},
	Comments:    make([]Post, 0),
	Attachments: make(Attachments, 0),
	Submitter:   &user.User{Uid: "188888131", Permission: 1},
}
