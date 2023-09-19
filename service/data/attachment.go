package data

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/cmps"
	"github.com/Drelf2018/request"
	"github.com/Drelf2020/utils"
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var log = utils.GetLog()
var attachmentCache = cmps.SafeSlice[*Attachment]{I: make([]*Attachment, 0)}

// 附件
type Attachment struct {
	Model
	// 原网址
	Url string `form:"url" gorm:"unique" cmps:"1"`
	// 本地路径
	Local string
	// 同浏览器 MIME type 附件的媒体类型
	MIME string
}

func (a Attachment) MarshalJSON() ([]byte, error) {
	return json.RawMessage(fmt.Sprintf(
		`{"url":"%v","local":"%v","MIME":"%v"}`,
		a.Url,
		utils.Ternary(a.Local != "", folder+a.Local, ""),
		a.MIME,
	)), nil
}

func (a Attachment) String() string {
	return fmt.Sprintf("Attachment(%v)", a.ID)
}

func (a *Attachment) Path() string {
	if a.Local == "" {
		a.Local = regexp.MustCompile("https?:/").ReplaceAllString(a.Url, "")
	}
	return a.Local
}

func (a *Attachment) Saved() bool {
	return Exists[Attachment]("url = ?", a.Url)
}

func (a *Attachment) Downloaded() bool {
	return Exists[Attachment]("url = ? and mime <> \"\"", a.Url)
}

func (a *Attachment) Store(data []byte) {
	dir, file := filepath.Split(a.Path())
	folder := public.MakeTo(dir)
	folder.MkdirAll()
	f, ok := folder.Touch(file, 0)
	if ok {
		f.Store(data)
	}
}

// 下载附件
func (a *Attachment) Download() error {
	result := request.Get(a.Url)
	if result.Error != nil {
		log.Errorf("Download %v error: %v", a, result.Error)
		return result.Error
	}
	// 判断完类型并保存在本地后再存数据库
	// 前两个操作都挺费时所以都协程了
	asyncio.Wait(
		asyncio.C(func() { a.MIME = mimetype.Detect(result.Content).String() }),
		asyncio.C(a.Store, result.Content),
	)
	asyncio.Retry(-1, 1, a.Saved)
	if err := db.Updates(a).Error; err != nil {
		log.Errorf("Update %v error: %v", a, err)
		return err
	}
	attachmentCache.Delete(a)
	return nil
}

func Save(url string) string {
	a := &Attachment{Url: url}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(a)
	asyncio.Retry(-1, 1, a.Downloaded)
	return a.Path()
}

func (a *Attachment) BeforeCreate(tx *gorm.DB) error {
	if Update(a, "url = ?", a.Url) {
		return nil
	}
	temp := attachmentCache.Search(a)
	if temp == nil || temp.ID == 0 {
		attachmentCache.Insert(a)
		go asyncio.RetryError(-1, 5, a.Download)
	} else {
		a.ID = temp.ID
	}
	return nil
}

// 附件合集
type Attachments []Attachment

func (as *Attachments) Add(urls ...string) {
	temp := make(Attachments, len(urls))
	for i, l := 0, len(urls); i < l; i++ {
		temp[i] = Attachment{Url: urls[i]}
	}
	*as = append(*as, temp...)
}

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
