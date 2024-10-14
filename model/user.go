package model

import (
	"time"

	"github.com/Drelf2018/webhook/model/serializer"
	"github.com/lib/pq"
	"gorm.io/gorm/schema"
)

const (
	Owner Permission = 1 << iota
	Admin
	Normal
)

type Permission uint64 // 权限

func (p Permission) Is(permissions ...Permission) bool {
	for _, v := range permissions {
		if p&v == 0 {
			return false
		}
	}
	return true
}

func (p Permission) Has(permissions ...Permission) bool {
	for _, v := range permissions {
		if p&v != 0 {
			return true
		}
	}
	return false
}

func (p Permission) IsTrusted() bool {
	return p.Has(Owner, Admin)
}

func (p Permission) IsOwner() bool {
	return p.Is(Owner)
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

func init() {
	schema.RegisterSerializer("error", serializer.ErrorSerializer)
}

// 请求记录
type RequestLog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"created_at"`
	BlogID    uint64    `json:"blog_id"`
	TaskID    uint64    `json:"task_id" gorm:"index:idx_logs_query"`
	MID       string    `json:"mid" gorm:"column:mid;index:idx_logs_query"`
	Type      string    `json:"type" gorm:"index:idx_logs_query"`
	Platform  string    `json:"platform" gorm:"index:idx_logs_query"`
	Result    []byte    `json:"result"`
	Error     error     `json:"error" gorm:"serializer:error"`
}

// 任务
type Task struct {
	ID          uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      string       `json:"-"`
	Name        string       `json:"name"`
	Enable      bool         `json:"enable"`
	Filter      Filter       `json:"filter" gorm:"embedded"`
	Api         Api          `json:"api" gorm:"embedded"`
	RequestOnce bool         `json:"request_once"`
	RequestLogs []RequestLog `json:"request_logs"`
}

// 用户
type User struct {
	Permission Permission `json:"permission"`
	BanTime    time.Time  `json:"ban_time"` // 封禁结束时间

	UID      string `json:"uid" gorm:"primaryKey"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"` // 鉴权码

	Tasks []Task `json:"tasks,omitempty"`
}
