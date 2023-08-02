package data

import (
	"database/sql/driver"
	"errors"

	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"gorm.io/gorm"
)

// 博文或评论
type Post struct {
	ID int64 `gorm:"primaryKey;autoIncrement" form:"-" json:"-"`
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
	Blogger
	// 附件
	Attachments `form:"-" json:"attachments"`
	// 回复
	Repost *Post `form:"-" json:"repost"`
	// 被回复
	Comments []*Comment `gorm:"-" form:"-" json:"comments"`
	// 提交者
	Submitter *user.User `form:"-"`
}

func (p *Post) Insert(c *Comment) {
	p.Comments = append(p.Comments, c)
}

// 保存该博文
func (p *Post) Save() {
	p.Content = utils.Clean(p.Text)
	p.Submitter.LevelUP()
	p.Blogger.Save()
	if p.Repost != nil && p.ID == 0 {
		p.Repost.Save()
	}
	db.Create(p)
}

// 保存分支
func (p *Post) SaveAsBranche() {
	db.Table("branches").Create(p)
}

// 获取分支
func GetBranches(platform, mid string, r *[]Post) {
	db.Table("branches").Find(r, "platform = ? AND mid = ?", platform, mid)
}

func (p *Post) Scan(val any) error {
	return db.First(p, "id = ?", val).Error
}

func (p Post) Value() (driver.Value, error) {
	return p.ID, nil
}

// 判断博文是否存在
func HasPost(platform, mid string) bool {
	return !errors.Is(db.First(new(Post), "platform = ? AND mid = ?", platform, mid).Error, gorm.ErrRecordNotFound)
}

// 通过平台和序号获取唯一博文
func GetPost(platform, mid string) *Post {
	var p Post
	db.First(&p, "platform = ? AND mid = ?", platform, mid)
	return &p
}

// 通过起始与结束时间获取范围内博文合集
func GetPosts(begin, end string, r *[]Post) {
	db.Find(&r, "time BETWEEN ? AND ?", begin, end)
}
