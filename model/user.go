package model

import (
	"time"

	"gorm.io/gorm"
)

// 用户
type User struct {
	UID       string    `json:"uid" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	IssuedAt  int64     `json:"-"`
	Ban       time.Time `json:"ban"`      // 封禁结束时间
	Role      Role      `json:"role"`     // 权限等级
	Name      string    `json:"name"`     // 用户名 非必要不可变
	Nickname  string    `json:"nickname"` // 昵称 可变
	Password  string    `json:"-"`        // 密码 不可变
	Tasks     []Task    `json:"tasks"`    // 任务集

	Extra map[string]any `json:"-" gorm:"serializer:json;->:false"` // 预留项 仅存
}

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

// 任务
type Task struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Public bool   `json:"public"`  // 是否公开
	Enable bool   `json:"enable"`  // 是否启用
	Name   string `json:"name"`    // 任务名称
	Method string `json:"method"`  // 请求方法
	URL    string `json:"url"`     // 请求地址
	Body   string `json:"body"`    // 请求内容
	Header Header `json:"header"`  // 请求头部
	README string `json:"README"`  // 任务描述
	ForkID uint64 `json:"fork_id"` // 复刻来源

	ForkCount int `json:"fork_count" gorm:"-"` // 被复刻次数

	Filters []Filter     `json:"filters"` // 筛选条件
	Logs    []RequestLog `json:"logs"`    // 请求记录
	UserID  string       `json:"user_id"` // 外键
}

func (t *Task) BeforeCreate(*gorm.DB) error {
	t.ID = 0
	t.CreatedAt = time.Time{}
	t.Logs = nil
	return nil
}

func (t *Task) AfterFind(tx *gorm.DB) error {
	return tx.Model(&Task{}).Select("count(*)").Find(&t.ForkCount, "fork_id = ?", t.ID).Error
}

// 博文筛选条件，用来描述一类博文，例如：
//
// filter1 表示所有平台为 "weibo"、类型为 "comment" 的博文
//
// filter2 表示所有由 "114" 提交的用户 "514" 的博文
//
//	var filter1 = Filter{
//		Platform: "weibo",
//		Type: "comment",
//	}
//
//	var filter2 = Filter{
//		Submitter: "114",
//		UID: "514",
//	}
type Filter struct {
	ID        uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	Submitter string `json:"submitter"` // 提交者
	Platform  string `json:"platform"`  // 发布平台
	Type      string `json:"type"`      // 博文类型
	UID       string `json:"uid"`       // 账户序号
	TaskID    uint64 `json:"-"`         // 外键
}

// 请求记录
type RequestLog struct {
	BlogID    uint64    `json:"blog_id"`
	CreatedAt time.Time `json:"created_at"`
	RawResult string    `json:"raw_result"`                    // 响应纯文本
	Result    any       `json:"result" gorm:"serializer:json"` // 响应为 JSON 会自动解析
	Error     string    `json:"error"`                         // 请求过程中发生的错误
	TaskID    uint64    `json:"-" gorm:"index:idx_logs_query"` // 外键
}
