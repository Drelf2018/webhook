package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/utils"
	"gopkg.in/yaml.v3"
)

// webhook 配置
type Config struct {
	Filename string         `toml:"-"      yaml:"-"      json:"-"`      // 配置文件保存路径
	Path     Path           `toml:"path"   yaml:"path"   json:"path"`   // 各文件路径
	Server   Server         `toml:"server" yaml:"server" json:"server"` // 服务器参数
	Role     Role           `toml:"role"   yaml:"role"   json:"role "`  // 权限组
	Extra    map[string]any `toml:"extra"  yaml:"extra"  json:"extra"`  // 额外项
}

func (c *Config) Import() (err error) {
	if c.Filename == "" {
		return
	}
	b, err := os.ReadFile(c.Filename)
	if err != nil {
		return
	}
	switch filepath.Ext(c.Filename) {
	case ".toml":
		err = toml.Unmarshal(b, c)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(b, c)
	case ".json":
		err = json.Unmarshal(b, c)
	}
	return
}

func (c *Config) Export() (err error) {
	if c.Filename == "" {
		return
	}
	err = os.MkdirAll(filepath.Dir(c.Filename), os.ModePerm)
	if err != nil {
		return
	}
	var b []byte
	switch filepath.Ext(c.Filename) {
	case ".toml":
		b, err = toml.Marshal(c)
	case ".yml", ".yaml":
		b, err = yaml.Marshal(c)
	case ".json":
		b, err = json.MarshalIndent(c, "", "  ")
	}
	if err != nil {
		return
	}
	return os.WriteFile(c.Filename, b, os.ModePerm)
}

type Path struct {
	Root   string `toml:"root"   yaml:"root"   json:"root"   default:"resource"`               // 程序根目录
	Logs   string `toml:"logs"   yaml:"logs"   json:"logs"   default:"logs"     join:"Root"`   // 记录文件夹
	UserDB string `toml:"userDB" yaml:"userDB" json:"userDB" default:"users.db" join:"Root"`   // 用户数据库文件
	BlogDB string `toml:"blogDB" yaml:"blogDB" json:"blogDB" default:"blogs.db" join:"Root"`   // 博文数据库文件
	Public string `toml:"public" yaml:"public" json:"public" default:"public"   join:"Root"`   // 公开文件夹
	Backup string `toml:"backup" yaml:"backup" json:"backup" default:"backup"   join:"Public"` // 博文数据库备份文件夹
	Upload string `toml:"upload" yaml:"upload" json:"upload" default:"upload"   join:"Public"` // 上传文件文件夹

	Full *Path `toml:"-" yaml:"-" json:"-"` // 以上字段的全路径 通过 utils.NewJoin 函数拼接
}

func (p *Path) copy() error {
	blogDB, err := os.Open(p.Full.BlogDB)
	if err != nil {
		return err
	}
	defer blogDB.Close()

	err = os.MkdirAll(p.Full.Backup, os.ModePerm)
	if err != nil {
		return err
	}
	file := filepath.Join(p.Full.Backup, time.Now().Format("2006-01-02.db"))
	backup, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer backup.Close()

	_, err = io.Copy(backup, blogDB)
	return err
}

func (p *Path) CopyBlogDB() {
	if err := p.copy(); err != nil {
		log := filepath.Join(p.Full.Backup, time.Now().Format("2006-01-02.log"))
		os.WriteFile(log, []byte(err.Error()), os.ModePerm)
	}
}

func (p *Path) AfterInitial() (err error) {
	p.Full, err = utils.NewJoin(*p)
	if err != nil {
		return
	}
	err = os.MkdirAll(p.Full.Logs, os.ModePerm)
	if err != nil {
		return
	}
	err = os.MkdirAll(p.Full.Backup, os.ModePerm)
	if err != nil {
		return
	}
	err = os.MkdirAll(p.Full.Upload, os.ModePerm)
	if err != nil {
		return
	}
	stop = time.AfterFunc(utils.NextTimeDuration(4, 0, 0), func() {
		p.CopyBlogDB()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-Quit.Done():
				return
			case <-ticker.C:
				go p.CopyBlogDB()
			}
		}
	}).Stop
	return
}

var _ initial.AfterInitial2 = (*Path)(nil)

type Server struct {
	Mode string `toml:"mode" yaml:"mode" json:"mode" default:"release"` // 模式
	Host string `toml:"host" yaml:"host" json:"host" default:"0.0.0.0"` // 主机
	Port uint16 `toml:"port" yaml:"port" json:"port" default:"9000"`    // 端口
}

func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Role struct {
	Owner string   `toml:"owner" yaml:"owner" json:"owner"` // 所有者
	Admin []string `toml:"admin" yaml:"admin" json:"admin"` // 管理员
}
