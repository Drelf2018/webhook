package model

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/webhook/utils"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/Drelf2018/intime"
)

var ErrSubmitted = errors.New("您已提交过")

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
	Avatar string `form:"avatar" json:"avatar"`
	// 装扮
	Pendant string `form:"pendant" json:"pendant"`
	// 头图
	Banner string `form:"banner" json:"banner"`
	// 博文序号
	Mid string `form:"mid" json:"mid"`
	// 发送时间
	Time intime.Time `form:"time" json:"time"`
	// 文本
	Text string `form:"text" json:"text"`
	// 来源
	Source string `form:"source" json:"source"`
	// 附件
	Attachments pq.StringArray `gorm:"type:text[]" form:"attachments" json:"attachments"`
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

const maxTextLength = 18

func (p *Post) Content() string {
	return utils.Clean(p.Text)
}

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

func (p *Post) ScanAndSend(rows *sql.Rows) {
	for rows.Next() {
		var job Job
		if rows.Scan(&job) != nil {
			continue
		}
		go job.Send(p)
	}
}

// 测试博文
var TestPost = &Post{
	Platform:    "weibo",
	Uid:         "7198559139",
	Name:        "七海Nana7mi",
	Avatar:      "https://wx4.sinaimg.cn/orj480/007Raq4zly8hd1vqpx3coj30u00u00uv.jpg",
	Follower:    "104.8万",
	Following:   "192",
	Description: "蓝色饭团",
	Mid:         "4952487292307646",
	Time:        1696248446000,
	Text:        `<span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span><span class="url-icon"><img alt=[赞] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_zan-44ddc70637.png" style="width:1em; height:1em;" /></span>`,
	Source:      "🦈iPhone 14 Pro Max",
	Submitter:   &User{Uid: "188888131", Permission: Owner},
	Repost: &Post{
		Platform:    "weibo",
		Uid:         "2203177060",
		Name:        "阿梓从小就很可爱",
		Avatar:      "https://wx4.sinaimg.cn/orj480/8351d064ly8hiph621dryj20u00u00vw.jpg",
		Follower:    "61.9万",
		Following:   "306",
		Description: "本人只喜欢读书",
		Mid:         "4952449691946355",
		Time:        1696239481000,
		Text:        "[看书] ",
		Source:      "iPhone 13 Pro Max",
		Attachments: pq.StringArray{"https://wx2.sinaimg.cn/large/8351d064ly1hih23pb486j21tk19k7wk.jpg"},
		Submitter:   &User{Uid: "188888131"},
		Repost:      &Post{},
	},
}
