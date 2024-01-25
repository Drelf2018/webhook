package model

import (
	"regexp"

	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/request"
	"gorm.io/gorm"
)

// 任务
type Job struct {
	gorms.Model[uint64]
	// 软删除
	Deleted gorm.DeletedAt `json:"-"`
	// 所有者
	UserUid string `json:"-"`
	// 监听样式
	Pattern string `form:"pattern" json:"pattern" yaml:"pattern"`
	// 任务参数
	request.Job
}

// 匹配任务
func (job *Job) Match(s string) bool {
	matched, err := regexp.MatchString(job.Pattern, s)
	return err == nil && matched
}

// 正则匹配任务
func GetJobsByRegexp(s string) (jobs []Job) {
	rows, _ := userDB.Model(&Job{}).Rows()
	defer rows.Close()
	for rows.Next() {
		var job Job
		rows.Scan(&job)
		if job.Match(s) {
			jobs = append(jobs, job)
		}
	}
	return
}

// 获取指定序号任务
func GetJobsByID(uid string, id []string) (jobs []Job) {
	userDB.Find(&jobs, "user_uid = ? and id IN ?", uid, id)
	return
}
