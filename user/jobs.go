package user

import (
	"database/sql/driver"
	"regexp"
	"strconv"
	"strings"

	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
)

type Job struct {
	// 数据库内序号
	ID     int64  `gorm:"primaryKey;autoIncrement" json:"-"`
	Patten string `form:"patten" json:"patten" yaml:"patten"`
	request.Job
}

// 转字符串
func (j Job) String() string {
	return strconv.Itoa(int(j.ID))
}

// 匹配
func (j Job) Match(s string) bool {
	matched, err := regexp.MatchString(j.Patten, s)
	if err == nil {
		return matched
	}
	return false
}

type Jobs []Job

// 转字符串
func (js *Jobs) ToString() string {
	s := utils.Await(Job.String, js)
	return strings.Join(s, ",")
}

func (js *Jobs) Scan(val any) error {
	return db.Where("id IN ?", strings.Split(val.(string), ",")).Find(js).Error
}

func (js Jobs) Value() (driver.Value, error) {
	return js.ToString(), nil
}

func GetJobsByRegexp(platform, uid string) (js Jobs) {
	var temp Jobs
	db.Where("patten LIKE ?", platform+"%").Find(&temp)
	for _, job := range temp {
		if job.Match(platform + uid) {
			js = append(js, job)
		}
	}
	return
}
