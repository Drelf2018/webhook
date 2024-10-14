package runner_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/model/runner"
	"github.com/glebarez/sqlite"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func ConvertTime(stamp int) time.Time {
	return time.Unix(int64(stamp/1000), int64(stamp%1000)*int64(time.Millisecond))
}

// 测试博文
var blog = &model.Blog{
	Submitter: "188888131",
	Platform:  "weibo",
	Type:      "blog",
	UID:       "7198559139",
	MID:       "4952487292307646",

	Text:   `<img alt=[good] src="https://h5.sinaimg.cn/m/emoticon/icon/others/h_good-0c51afc69c.png" style="width:1em; height:1em;" />`,
	Time:   ConvertTime(1696248446000),
	Source: "🦈iPhone 14 Pro Max",

	Name:        "七海Nana7mi",
	Avatar:      "https://wx4.sinaimg.cn/orj480/007Raq4zly8hd1vqpx3coj30u00u00uv.jpg",
	Follower:    "104.8万",
	Following:   "192",
	Description: "蓝色饭团",

	Reply: &model.Blog{
		Platform: "weibo",
		Type:     "blog",
		UID:      "2203177060",
		MID:      "4952449691946355",

		Text:   "[看书] ",
		Time:   ConvertTime(1696239481000),
		Source: "iPhone 13 Pro Max",

		Name:        "阿梓从小就很可爱",
		Avatar:      "https://wx4.sinaimg.cn/orj480/8351d064ly8hiph621dryj20u00u00vw.jpg",
		Follower:    "61.9万",
		Following:   "306",
		Description: "本人只喜欢读书",

		Assets: pq.StringArray{"https://wx2.sinaimg.cn/large/8351d064ly1hih23pb486j21tk19k7wk.jpg"},
	},

	Extra: map[string]any{
		"$like": "1k+",
	},
}

func TestCreateBlog(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("blogs.db"))
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&model.Blog{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Create(blog).Error
	if err != nil {
		t.Fatal(err)
	}
}

var logsDB *gorm.DB

func init() {
	var err error
	// logsDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	logsDB, err = gorm.Open(sqlite.Open("logs.db"))
	if err != nil {
		panic(err)
	}

	err = logsDB.AutoMigrate(&model.User{}, &model.Task{}, &model.RequestLog{})
	if err != nil {
		panic(err)
	}

	err = logsDB.Create([]model.Task{
		{
			Enable:      true,
			RequestOnce: false,
			Filter: model.Filter{
				Platform: []string{"weibo"},
			},
			Api: model.Api{
				Method: http.MethodPost,
				URL:    "https://httpbin.org/anything/{{.UID}}",
				Body:   `"{{.Name}}：“{{.Text}}”"`,
				Header: model.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			Enable:      true,
			RequestOnce: true,
		},
		{
			Enable:      false,
			RequestOnce: false,
		},
		{
			Enable:      true,
			RequestOnce: true,
			Api: model.Api{
				Method:    http.MethodPost,
				URL:       "https://httpbin.org/anything/{{.UID}}",
				Header:    model.Header{"Content-Type": []string{"application/json"}},
				Body:      `{{json .}}`,
				Parameter: []string{model.DoNotUnmarshal},
			},
		},
	}).Error
	if err != nil {
		panic(err)
	}

	err = logsDB.Create([]model.RequestLog{
		{
			TaskID:   2,
			Platform: blog.Platform,
			Type:     blog.Type,
			MID:      blog.MID,
		},
		{
			TaskID:   3,
			Platform: blog.Platform,
			Type:     blog.Type,
			MID:      blog.MID,
		},
	}).Error
	if err != nil {
		panic(err)
	}
}

func TestFilter(t *testing.T) {
	runner := runner.TaskRunner{DB: logsDB, Timeout: 10 * time.Second}
	runner.SendBlog(blog)
}
