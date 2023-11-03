package data

import (
	"bytes"
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

// 附件
type Attachment struct {
	db.Model
	// 原网址
	Url string `form:"url" gorm:"unique;not null" cmps:"1"`
	// 本地路径
	Local string
	// 同浏览器 MIME type 附件的媒体类型
	MIME string
}

func (a Attachment) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`{"url":"`)
	b.WriteString(a.Url)
	b.WriteString(`","local":"`)
	if a.Path() != "" {
		b.WriteString(folder)
		b.WriteString(a.Path())
	}
	b.WriteString(`","MIME":"`)
	b.WriteString(a.MIME)
	b.WriteString(`"}`)
	return json.RawMessage(b.Bytes()), nil
}

func (a Attachment) String() string {
	return fmt.Sprintf("Attachment(%v, %v%v)", a.ID, a.Url, utils.Ternary(a.MIME == "", "", ", "+a.MIME))
}

func (a *Attachment) Path() string {
	if a.Local == "" {
		a.Local = regexp.MustCompile("https?:/").ReplaceAllString(a.Url, "")
	}
	return a.Local
}

func (a *Attachment) Save(_ any) {
	if a.Url == "" {
		return
	}
	Data.FirstOrCreate(nil, func() { go asyncio.RetryError(-1, 5, a.Download) }, a, "url = ?", a.Url)
}

func (a *Attachment) Detect(r *request.Result) {
	a.MIME = mimetype.Detect(r.Content).String()
}

func (a *Attachment) Store(r *request.Result) {
	dir, _ := filepath.Split(a.Path())
	dir = filepath.Join(public, dir)
	os.MkdirAll(dir, os.ModePerm)
	r.Write(filepath.Join(public, a.Path()), os.ModePerm)
}

// 下载附件
func (a *Attachment) Download() error {
	if a.Url == "" {
		return nil
	}
	result := request.Get(a.Url, request.Referer("https://weibo.com/"))
	if utils.LogErr(result.Error()) {
		return result.Error()
	}
	// 前两个操作都挺费时所以都协程了
	asyncio.ForFunc(result, a.Detect, a.Store)
	// 判断完类型并保存在本地后再存数据库
	if err := Data.DB.Updates(a).Error; utils.LogErr(err) {
		return err
	}
	return nil
}

func Save(url string) string {
	a := &Attachment{Url: url}
	Data.FirstOrCreate(nil, func() { asyncio.RetryError(-1, 5, a.Download) }, a, "url = ?", url)
	return folder + a.Path()
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
	aList := make([]string, 0, len(as))
	asyncio.ForEach(as, func(a Attachment) { aList = append(aList, a.String()) })
	return fmt.Sprintf("[%v]", strings.Join(aList, ", "))
}
