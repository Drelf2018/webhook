package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/service/db"
	"github.com/Drelf2020/utils"
	"github.com/gabriel-vasile/mimetype"
)

var https = regexp.MustCompile(`https?:/`)

// 附件
type Attachment struct {
	db.Model
	// 原网址
	Url string `json:"url" form:"url" gorm:"unique;not null" cmps:"1"`
	// 本站链接
	Link string `json:"link"`
	// 本地路径
	Local string `json:"-"`
	// 同浏览器 MIME type 附件的媒体类型
	MIME string `json:"MIME"`
}

func (a *Attachment) Init() {
	p := https.ReplaceAllString(a.Url, "")
	a.Link = folder + p
	a.Local = filepath.Join(public, p)
}

func (a Attachment) String() string {
	return fmt.Sprintf("Attachment(%v, %v)", a.ID, a.Url)
}

func (a *Attachment) Save(_ any) {
	if a.Url == "" {
		return
	}
	Posts.FirstOrCreate(nil, func() { go asyncio.RetryError(-1, 5, a.Download) }, a, "url = ?", a.Url)
}

func (a *Attachment) Detect(content []byte) {
	a.MIME = mimetype.Detect(content).String()
}

func (a *Attachment) Store(content []byte) {
	dir, _ := filepath.Split(a.Local)
	os.MkdirAll(dir, os.ModePerm)
	os.WriteFile(a.Local, content, os.ModePerm)
}

// 下载附件
func (a *Attachment) Download() error {
	if a.Url == "" {
		return nil
	}
	a.Init()

	// request
	result := request.Get(a.Url, func(job *request.Job) {
		job.Headers = request.HEADERS
		if strings.Contains(a.Url, "sinaimg.cn") {
			job.Headers["Referer"] = "https://weibo.com/"
		}
	})
	if utils.LogErr(result.Error()) {
		return result.Error()
	}

	// 前两个操作都挺费时所以都协程了
	asyncio.ForFunc(result.Content, a.Detect, a.Store)

	// 判断完类型并保存在本地后再存数据库
	if err := Posts.DB.Updates(a).Error; utils.LogErr(err) {
		return err
	}
	return nil
}

func Save(url string) string {
	a := &Attachment{Url: url}
	Posts.FirstOrCreate(nil, func() { asyncio.RetryError(-1, 5, a.Download) }, a, "url = ?", url)
	return a.Link
}

// 附件合集
type Attachments []Attachment

func (as Attachments) Urls() string {
	temp := make([]string, len(as))
	for i, l := 0, len(as); i < l; i++ {
		temp[i] = as[i].Url
	}
	b, err := json.Marshal(temp)
	if err != nil {
		return ""
	}
	return string(b)
}

func (as Attachments) String() string {
	return fmt.Sprintf("[%v]", strings.Join(asyncio.ForEachV(as, Attachment.String), ", "))
}
