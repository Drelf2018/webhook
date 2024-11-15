package model

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// 标签
type Tag struct {
	Tag  string `json:"tag"`  // 文字描述
	Href string `json:"href"` // 引用网址
}

// 博文
type Blog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"created_at"`

	Submitter string `json:"submitter" gorm:"index:idx_blogs_query,priority:2"`      // 提交者
	Platform  string `json:"platform" gorm:"index:idx_blogs_query,priority:5"`       // 发布平台
	Type      string `json:"type" gorm:"index:idx_blogs_query,priority:4"`           // 博文类型
	UID       string `json:"uid" gorm:"index:idx_blogs_query,priority:3"`            // 账户序号
	MID       string `json:"mid" gorm:"index:idx_blogs_query,priority:1;column:mid"` // 博文序号

	URL    string    `json:"url"`    // 博文网址
	Text   string    `json:"text"`   // 文本内容
	Time   time.Time `json:"time"`   // 发送时间
	Source string    `json:"source"` // 博文来源
	Edited bool      `json:"edited"` // 是否编辑

	Name        string `json:"name"`        // 账户昵称
	Avatar      string `json:"avatar"`      // 头像网址
	Follower    string `json:"follower"`    // 粉丝数
	Following   string `json:"following"`   // 关注数
	Description string `json:"description"` // 个人简介

	ReplyID  *uint64 `json:"reply_id"`   // 被回复博文序号
	Reply    *Blog   `json:"reply"`      // 被回复博文
	BlogID   *uint64 `json:"comment_id"` // 被评论博文序号
	Comments []Blog  `json:"comments"`   // 本博文的评论

	Assets pq.StringArray `json:"assets" gorm:"type:text[]"` // 资源网址
	Banner pq.StringArray `json:"banner" gorm:"type:text[]"` // 头图网址

	Tags  []Tag          `json:"tags" gorm:"serializer:json"`  // 标签
	Extra map[string]any `json:"extra" gorm:"serializer:json"` // 预留项
}

func (b *Blog) BeforeCreate(tx *gorm.DB) error {
	b.ID = 0
	b.CreatedAt = time.Time{}
	if b.Reply != nil {
		b.Reply.Submitter = b.Submitter
	}
	b.Comments = nil
	return nil
}

var MaxTextLength = 18

func (b Blog) String() string {
	text := b.Text
	if len(text) > MaxTextLength {
		text = text[:MaxTextLength] + "..."
	}
	if b.Reply == nil {
		return fmt.Sprintf("Blog(%d, %s, %s)", b.ID, b.Name, text)
	}
	return fmt.Sprintf("Blog(%d, %s, %s, %s)", b.ID, b.Name, text, b.Reply)
}
