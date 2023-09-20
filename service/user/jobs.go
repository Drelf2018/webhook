package user

import (
	"regexp"

	"github.com/Drelf2018/request"
	"github.com/Drelf2020/utils"
)

// 回调任务封装
type Job struct {
	// 数据库内序号
	ID      int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUid string `json:"-"`
	Patten  string `form:"patten" json:"patten" yaml:"patten"`
	request.Job
}

// 匹配
func (j Job) Match(s string) bool {
	matched, err := regexp.MatchString(j.Patten, s)
	return err == nil && matched
}

// 正则匹配任务
func GetJobsByRegexp(platform, uid string) []Job {
	var temp []Job
	db.Where("patten LIKE ?", platform+"%").Find(&temp)
	return utils.Filter(temp, func(j Job) bool { return j.Match(platform + uid) })
}
