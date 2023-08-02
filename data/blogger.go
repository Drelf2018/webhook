package data

import (
	"database/sql/driver"
	"time"
)

// 博主信息部分
type Blogger struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" form:"-" json:"-"`
	Platform    string     `form:"platform" json:"platform"`
	Uid         string     `form:"uid" json:"uid"`
	CreatedAt   string     `gorm:"column:create" form:"create" json:"create"`
	Name        string     `form:"name" json:"name"`
	Face        Attachment `form:"-" json:"face"`
	Pendant     Attachment `form:"-" json:"pendant"`
	Description string     `form:"description" json:"description"`
	Follower    string     `form:"follower" json:"follower"`
	Following   string     `form:"following" json:"following"`
}

func (b *Blogger) Save() {
	if b.ID != 0 {
		return
	}
	db.Create(b)
}

// 查询某一时刻前最近的用户状态
func (b *Blogger) Query(now time.Time) *Blogger {
	db.Last(b, "`create` <= ?", now)
	return b
}

func (b *Blogger) Scan(val any) error {
	return db.First(b, "id = ?", val).Error
}

func (b Blogger) Value() (driver.Value, error) {
	return b.ID, nil
}
