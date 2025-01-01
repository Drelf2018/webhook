package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var BaseURL string = ""

var tmpl, _ = template.New("rss").Funcs(template.FuncMap{
	"download": func(url string) string {
		return fmt.Sprintf("%s/public/%s", BaseURL, strings.ReplaceAll(url, ":/", ""))
	},
}).Parse(`
<div style="width: max-content;margin-top: 0.5em; display: flex;padding: 0.5em;align-items: center;border-radius:1em;box-shadow: 0 3px 1px -2px #0000001f, 0 2px 2px #00000024, 0 1px 5px #0003;">
    <div style="background-repeat: round;width: 4rem;background-image: url({{ download .Avatar }});height: 4rem;border-radius: 1em;box-shadow: 0 3px 1px -2px #0000001f, 0 2px 2px #00000024, 0 1px 5px #0003;margin-right: 0.5em;"></div>
    <div style="margin-right: 0.5em;">
        <div style="font-size: 1.5em;color: #343233;font-weight: bold;">{{ .Author }}</div>
        <div style="color: grey;">{{ .Description }}</div>
    </div>
</div>`)

func GetUserInfo(item *Item) string {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, item)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return ""
	}
	return buf.String()
}

type Asset struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func NewAsset(url string) (a Asset) {
	a.URL = fmt.Sprintf("%s/public/%s", BaseURL, strings.ReplaceAll(url, ":/", ""))
	// download
	httpFile, err := downloader.Download(url)
	if err != nil {
		return
	}
	// Length
	if file, ok := httpFile.(*os.File); ok {
		if info, err := file.Stat(); err == nil {
			a.Length = info.Size()
		}
	}
	// Type
	b := make([]byte, 512)
	if _, err = httpFile.Read(b); err == nil {
		a.Type = http.DetectContentType(b)
	}
	return
}

type Assets []Asset

func (a *Assets) Scan(src any) error {
	var r pq.StringArray
	err := r.Scan(src)
	if err != nil {
		return err
	}
	for _, url := range r {
		*a = append(*a, NewAsset(url))
	}
	return nil
}

type GUID uint64

func (id GUID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(fmt.Sprintf("%s/blog/%d", BaseURL, id), start)
}

type PubDate time.Time

func (t PubDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(time.Time(t).Format("2006-01-02 15:04:05"), start)
}

type Item struct {
	Title       string  `xml:"title"`                          // 文章标题
	Link        string  `xml:"link" gorm:"column:url"`         // 博文网址
	Desc        string  `xml:"description" gorm:"column:text"` // 文本内容
	Author      string  `xml:"author" gorm:"column:name"`      // 账户昵称
	Assets      Assets  `xml:"enclosure" gorm:"type:text[]"`   // 博文附件
	GUID        GUID    `xml:"guid" gorm:"column:id"`          // 固定链接
	PubDate     PubDate `xml:"pubDate" gorm:"column:time"`     // 发送时间
	Source      string  `xml:"source"`                         // 博文来源
	Avatar      string  `xml:"-"`
	Description string  `xml:"-"`
}

type Channel struct {
	Title     string  `xml:"title" gorm:"column:name"`
	Link      string  `xml:"link" gorm:"-"`
	Desc      string  `xml:"description" gorm:"column:README"`
	Language  string  `xml:"language" gorm:"-"`
	Copyright string  `xml:"copyright" gorm:"-"`
	Generator string  `xml:"generator" gorm:"-"`
	Items     []*Item `xml:"item" gorm:"-"`

	ID        uint64         `xml:"-"`
	CreatedAt time.Time      `xml:"-"`
	UserID    string         `xml:"-"`                          // 外键
	Filters   []model.Filter `xml:"-" gorm:"foreignKey:TaskID"` // 筛选条件
}

func (c *Channel) AfterFind(tx *gorm.DB) error {
	// 初始化数据
	c.Link = fmt.Sprintf("%s/task/%d", BaseURL, c.ID)
	c.Language = "zh-CN"
	tx.Model(&model.User{}).Select("name").Find(&c.Copyright, "uid = ?", c.UserID)
	now := time.Now().Year()
	create := c.CreatedAt.Year()
	if now == create {
		c.Copyright = fmt.Sprintf("Copyright %d, %s", create, c.Copyright)
	} else {
		c.Copyright = fmt.Sprintf("Copyright %d-%d, %s", create, now, c.Copyright)
	}
	_, c.Generator, _ = strings.Cut(BaseURL, "://")
	// 获取博文
	filter := BlogDB.Model(&model.Blog{})
	for _, f := range c.Filters {
		f.TaskID = 0
		filter = filter.Or(f)
	}
	err := BlogDB.Model(&model.Blog{}).Where(filter).Find(&c.Items).Error
	if err != nil {
		return err
	}
	for _, item := range c.Items {
		if item.Title == "" {
			item.Title = item.Author
		}
		for _, asset := range item.Assets {
			if strings.HasPrefix(asset.Type, "image") {
				item.Desc += fmt.Sprintf("<br /><img src=\"%s\" referrerpolicy=\"no-referrer\" style=\"vertical-align:middle;\" />", asset.URL)
			} else if strings.HasPrefix(asset.Type, "video") {
				item.Desc += fmt.Sprintf("<br /><video controls><source src=\"%s\" type=\"%s\"></video>", asset.URL, asset.Type)
			}
		}
		item.Desc += GetUserInfo(item)
	}
	return nil
}

type rss struct {
	Version string  `xml:"version,attr"`
	Channel Channel `xml:"channel"`
}

func GetRssID(ctx *gin.Context) (any, error) {
	r := rss{Version: "2.0"}
	uid, _ := JWTAuth(ctx)
	tx := UserDB.Model(&model.Task{}).Preload("Filters").Limit(1).Find(&r.Channel, "id = ? AND (public OR user_id = ?)", ctx.Param("id"), uid)
	if tx.Error != nil {
		return 1, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 2, ErrTaskNotExist
	}
	b, err := xml.MarshalIndent(r, "", "  ")
	if err != nil {
		return 3, err
	}
	ctx.Writer.WriteString(xml.Header)
	ctx.Writer.Write(b)
	return nil, nil
}
