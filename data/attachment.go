package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
)

// 附件
type Attachment struct {
	// 数据库内序号
	ID int64 `gorm:"primaryKey;autoIncrement" json:"-"`
	// 同浏览器 MIME type 附件的媒体类型
	MIME string
	// 本地路径
	Path string `json:"path"`
	// 数据
	data []byte
}

// 构造函数
func (a *Attachment) Make(url string) string {
	if url != "" {
		a.Path = regexp.MustCompile("https?:/").ReplaceAllString(url, "")
		if a.NotExist() {
			a.Download(url)
		}
	}
	return a.Path
}

func (a *Attachment) Scan(val any) error {
	if val.(int64) == 0 {
		return nil
	}
	return db.First(a, "id = ?", val).Error
}

func (a Attachment) Value() (driver.Value, error) {
	return a.ID, nil
}

// 转字符串
func (a Attachment) String() string {
	return strconv.Itoa(int(a.ID))
}

// 转链接
func (a Attachment) ToURL() string {
	return Resource + a.Path
}

// 判断附件是否存在
func (a *Attachment) NotExist() bool {
	return errors.Is(db.First(a, "path = ?", a.Path).Error, gorm.ErrRecordNotFound)
}

// 判断格式
func (a *Attachment) ParseType() {
	a.MIME = mimetype.Detect(a.data).String()
}

// 保存到本地
func (a *Attachment) SaveToLocal() {
	os.MkdirAll(Resource+filepath.Dir(a.Path), os.ModePerm)
	os.WriteFile(Resource+a.Path, a.data, os.ModePerm)
}

// 下载附件
func (a *Attachment) Download(url string) {
	result := request.Get(url)
	if result == nil {
		return
	}
	a.data = result.Data

	// 判断完类型并保存在本地后再存数据库
	// 然后前两个操作都挺费时所以都协程了
	utils.All(a.ParseType, a.SaveToLocal)
	db.Create(a)
}

// 附件合集
type Attachments []Attachment

func (as *Attachments) Make(urls ...string) {
	*as = make(Attachments, len(urls))
	utils.List(func(i int) { (*as)[i].Make(urls[i]) }, len(urls))
}

// 转字符串
func (as *Attachments) ToString() string {
	s := utils.Await(Attachment.String, as)
	return strings.Join(s, ",")
}

// 转链接
func (as *Attachments) ToURL() string {
	l := make([]string, len(*as))
	for i, a := range *as {
		l[i] = a.ToURL()
	}
	b, _ := json.Marshal(l)
	return string(b)
}

func (as *Attachments) Scan(val any) error {
	return db.Where("id IN ?", strings.Split(val.(string), ",")).Find(as).Error
}

func (as Attachments) Value() (driver.Value, error) {
	return as.ToString(), nil
}
