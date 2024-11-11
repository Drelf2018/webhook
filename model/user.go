package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	Invalid Role = iota
	Normal
	Trusted
	Admin
	Owner
)

type Role uint64 // 权限

func (r Role) IsAdmin() bool {
	return r == Owner || r == Admin
}

const (
	FilterQuery string = "enable AND (submitter IS NULL OR submitter = '' OR submitter = '{}' OR submitter LIKE '%%%s%%') AND (uid IS NULL OR uid = '' OR uid = '{}' OR uid LIKE '%%%s%%') AND (type IS NULL OR type = '' OR type = '{}' OR type LIKE '%%%s%%') AND (platform IS NULL OR platform = '' OR platform = '{}' OR platform LIKE '%%%s%%') "
	ExceptQuery string = "request_once AND EXISTS (SELECT 1 FROM request_logs WHERE task_id = tasks.id AND mid = ? AND type = ? AND platform = ? LIMIT 1)"
)

type Filter struct {
	Submitter pq.StringArray `json:"submitter" gorm:"type:text[]"`
	Platform  pq.StringArray `json:"platform" gorm:"type:text[]"`
	Type      pq.StringArray `json:"type" gorm:"type:text[]"`
	UID       pq.StringArray `json:"uid" gorm:"type:text[]"`
}

// 请求记录
type RequestLog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"created_at"`
	TaskID    uint64    `json:"task_id" gorm:"index:idx_logs_query"`
	BlogID    uint64    `json:"blog_id"`
	MID       string    `json:"mid" gorm:"index:idx_logs_query;column:mid"`
	Type      string    `json:"type" gorm:"index:idx_logs_query"`
	Platform  string    `json:"platform" gorm:"index:idx_logs_query"`
	Result    string    `json:"result"`
	Error     string    `json:"error"`
}

// 任务
type Task struct {
	ID          uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Enable      bool           `json:"enable"`
	UserID      string         `json:"user_id"`
	Name        string         `json:"name"`
	Filter      Filter         `json:"filter" gorm:"embedded"`
	Api         Api            `json:"api" gorm:"embedded"`
	RequestOnce bool           `json:"request_once"`
	RequestLogs []RequestLog   `json:"request_logs,omitempty"`
}

// 用户
type User struct {
	Role     Role           `json:"role"` // 权限等级
	Ban      time.Time      `json:"ban"`  // 封禁结束时间
	UID      string         `json:"uid" gorm:"primaryKey"`
	Name     string         `json:"name"`
	Nickname string         `json:"nickname"` // 权限昵称
	Password string         `json:"-"`
	Tasks    []Task         `json:"tasks"`
	Extra    map[string]any `json:"-" gorm:"serializer:json"` // 预留项
}
