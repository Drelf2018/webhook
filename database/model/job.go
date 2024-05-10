package model

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/request"
	"gorm.io/gorm"
)

func MatchRegexp(pattern, text string) bool {
	matched, err := regexp.MatchString(pattern, text)
	return err == nil && matched
}

type Filter struct {
	// Specific users and platforms
	Uid      string `form:"uid"         json:"uid"         yaml:"uid"`
	Platform string `form:"platform"    json:"platform"    yaml:"platform"`
	// Accept posts from all sources
	AllSources bool `form:"all_sources" json:"all_sources" yaml:"all_sources"`
}

func (f Filter) Match(p *Post) bool {
	if !f.AllSources && !p.Submitter.IsTrusted() {
		return false
	}
	if f.Platform != "" && !MatchRegexp(f.Platform, p.Platform) {
		return false
	}
	if f.Uid != "" && !MatchRegexp(f.Uid, p.Uid) {
		return false
	}
	return true
}

// 任务
type Job struct {
	gorms.Model[uint64]
	// 过滤器
	Filter
	// 任务参数
	request.Job
	// 所有者
	UserUid string `json:"-"`
	// 软删除
	Deleted gorm.DeletedAt `json:"-"`
}

func (job *Job) Template() (*template.Template, error) {
	b, err := json.Marshal(job.Job)
	if err != nil {
		return nil, err
	}

	return template.New("").Parse(strings.ReplaceAll(string(b), `\"`, `"`))
}

func (job *Job) Execute(p *Post) error {
	tmpl, err := job.Template()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, p)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf.Bytes(), &job.Job)
}

func (job *Job) Send(p *Post) *request.Response {
	if !job.Match(p) {
		return nil
	}

	err := job.Execute(p)
	if err != nil {
		return request.NewError(err)
	}

	return job.Do()
}

var TestJob = &Job{
	Filter: Filter{
		Platform: "weibo",
		Uid:      "\\d+",
	},
	Job: request.Job{
		Method: http.MethodGet,
		Url:    "http://api.nana7mi.link:5760/weibo",
		Query: request.M{
			"channel": "9673211",
			"message": `{{.Name}}
粉丝 {{.Follower}} | 关注 {{.Following}}

“{{.Content}}”
{{if .Repost}}
  {{.Repost.Name}}
  粉丝 {{.Repost.Follower}} | 关注 {{.Repost.Following}}
  
  “{{.Repost.Content}}”

  {{.Repost.Time.Format "%m-%d %H:%M:%S"}}
  来自{{.Repost.Source}}
{{end}}
{{.Time.Format "%m-%d %H:%M:%S"}}
来自{{.Source}}`,
		},
	},
}
