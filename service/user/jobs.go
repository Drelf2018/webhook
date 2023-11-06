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
	Pattern string `form:"pattern" json:"pattern" yaml:"pattern"`
	request.Job
}

// 匹配
func (job Job) Match(s string) bool {
	matched, err := regexp.MatchString(job.Pattern, s)
	return err == nil && matched
}

// 正则匹配任务
func GetJobsByRegexp(platform, uid string) []Job {
	var jobs []Job
	Users.Find(&jobs, "pattern LIKE ?", platform+"%")
	return utils.Filter(jobs, func(job Job) bool { return job.Match(platform + uid) })
}

// 获取指定序号任务
func GetJobsByID(uid string, id ...string) (jobs []Job) {
	Users.Find(&jobs, "user_uid = ? and id IN ?", uid, id)
	return
}
