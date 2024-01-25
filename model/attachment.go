package model

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Drelf2018/request"
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
)

var (
	reg     = regexp.MustCompile(`/?https?:/`)
	normal  = request.New(http.MethodGet, "", request.Headers(request.HEADERS))
	sinaimg = request.New(http.MethodGet, "", request.Referer("https://weibo.com/"))
)

func CleanURL(url string) string {
	return reg.ReplaceAllString(url, "")
}

func ConvertURL(url string) string {
	return filepath.Clean(reg.ReplaceAllString(url, `public\`))
}

type Attachment struct {
	Url  string `gorm:"primaryKey" json:"url"`
	MIME string
}

func (a *Attachment) AfterCreate(tx *gorm.DB) error {
	go a.MustDownload(3)
	return nil
}

func (a *Attachment) Get() *request.Result {
	if strings.Contains(a.Url, "sinaimg.cn") {
		return sinaimg.Fetch(a.Url)
	}
	return normal.Fetch(a.Url)
}

func (a *Attachment) Download() (err error) {
	if a.Url == "" {
		return nil
	}
	// fetch
	result := a.Get()
	err = result.Error()
	if err != nil {
		return
	}
	// store
	err = result.Write(ConvertURL(a.Url), os.ModePerm)
	if err != nil {
		return
	}
	// update
	a.MIME = mimetype.Detect(result.Content).String()
	return postDB.Updates(a).Error
}

func (a *Attachment) MustDownload(sleep float64) {
	for a.Download() != nil {
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

// 下载附件
func Download(url string) (a *Attachment) {
	a = &Attachment{Url: url}
	a.MustDownload(3)
	return
}
