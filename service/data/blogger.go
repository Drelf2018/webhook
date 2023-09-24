package data

import (
	"fmt"
	"time"

	"github.com/Drelf2018/webhook/service/db"
)

// 博主信息部分
type Blogger struct {
	db.Model
	Platform    string `form:"platform" json:"platform"`
	Uid         string `form:"uid" json:"uid" cmps:"1"`
	Time        string `form:"time" json:"time"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"description"`
	Follower    string `form:"follower" json:"follower"`
	Following   string `form:"following" json:"following"`

	FaceID int64      `gorm:"column:face" json:"-"`
	Face   Attachment `form:"face" json:"face" preload:"1"`

	PendantID int64      `gorm:"column:pendant" json:"-"`
	Pendant   Attachment `form:"pendant" json:"pendant" preload:"2"`
}

func (b Blogger) String() string {
	return fmt.Sprintf("Blogger(id=%v, platform=%v, uid=%v, name=%v, face=%v, pendant=%v)", b.ID, b.Platform, b.Uid, b.Name, b.Face, b.Pendant)
}

// 查询某一时刻前最近的用户状态
func (b *Blogger) Query(now time.Time) *Blogger {
	Data.DB.Last(b, "`create` <= ?", now)
	return b
}
