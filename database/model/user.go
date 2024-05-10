package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var ErrNoSubmitter = errors.New("model: post has no submitter")
var ErrType = errors.New("model: failed to unmarshal JSONB value")

// 提交者
type User struct {
	Permission `json:"permission"`
	// 哔哩哔哩 uid
	Uid string `json:"uid" gorm:"primaryKey"`
	// 鉴权码
	Auth string `json:"auth,omitempty"`
	// 账户下任务
	Jobs []Job `json:"jobs,omitempty" gorm:"foreignKey:UserUid"`
	// 关注的账号
	Follow pq.StringArray `gorm:"type:text[]" json:"follow,omitempty"`
}

func (u User) String() string {
	return fmt.Sprintf("%v@%s", u.Permission, u.Uid)
}

func (u *User) Scan(val any) error {
	if val == nil {
		*u = User{}
		return nil
	}
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return ErrType
	}
	return json.Unmarshal(ba, u)
}

func (u *User) Value() (driver.Value, error) {
	if u == nil {
		return nil, ErrNoSubmitter
	}
	return fmt.Sprintf(`{"uid":"%s","permission":%d}`, u.Uid, u.Permission), nil
}
