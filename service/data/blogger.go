package data

import (
	"fmt"
	"time"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/cmps"
	"gorm.io/gorm"
)

var bloggerCache = cmps.SafeSlice[*Blogger]{I: make([]*Blogger, 0)}

// 博主信息部分
type Blogger struct {
	Model
	Platform    string `form:"platform" json:"platform"`
	Uid         string `form:"uid" json:"uid" cmps:"1"`
	Create      string `form:"create" json:"create"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"description"`
	Follower    string `form:"follower" json:"follower"`
	Following   string `form:"following" json:"following"`

	FaceID int64      `gorm:"column:face" json:"-"`
	Face   Attachment `form:"face" json:"face" preload:"1"`

	PendantID int64      `gorm:"column:pendant" json:"-"`
	Pendant   Attachment `form:"pendant" json:"pendant" preload:"2"`
}

func (b *Blogger) Saved() bool {
	return Exists[Blogger](b)
}

func (b *Blogger) BeforeCreate(tx *gorm.DB) error {
	if Update(b, b) {
		return nil
	}
	temp := bloggerCache.Search(b)
	if temp == nil || temp.ID == 0 {
		bloggerCache.Insert(b)
		go func() {
			asyncio.Retry(-1, 5, b.Saved)
			bloggerCache.Delete(b)
		}()
	} else {
		b.ID = temp.ID
	}
	return nil
}

func (b Blogger) String() string {
	return fmt.Sprintf("Blogger(id=%v, platform=%v, uid=%v, name=%v, face=%v, pendant=%v)", b.ID, b.Platform, b.Uid, b.Name, b.Face, b.Pendant)
}

// 查询某一时刻前最近的用户状态
func (b *Blogger) Query(now time.Time) *Blogger {
	db.Last(b, "`create` <= ?", now)
	return b
}
