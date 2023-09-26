package configs

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/resource"
	"github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

// 设置值类型对象默认值
func SetZero[T comparable](a *T, b ...T) {
	var zero T
	if *a == zero {
		for _, c := range b {
			if c == zero {
				continue
			}
			*a = c
			break
		}
	}
}

// 设置引用类型对象默认值
func SetNil[C any, T utils.CanNil[C]](a *T, b T) {
	if *a == nil {
		*a = b
	}
}

// 资源文件夹路径
type Path struct {
	Root   string `yaml:"root"`
	Views  string `yaml:"views"`
	Public string `yaml:"public"`
	Posts  string `yaml:"posts"`
	Users  string `yaml:"users"`
	Log    string `yaml:"log"`
	Full   struct {
		Views  string
		Public string
		Posts  string
		Users  string
		Log    string

		Index   string
		Version string
	} `yaml:"-"`
}

func (p *Path) Init() {
	SetZero(&p.Root, "resource")
	SetZero(&p.Views, "views")
	SetZero(&p.Public, "public")
	SetZero(&p.Posts, "posts.db")
	SetZero(&p.Users, "users.db")
	SetZero(&p.Log, ".log")

	p.Full.Views = filepath.Join(p.Root, p.Views)
	p.Full.Public = filepath.Join(p.Root, p.Public)
	p.Full.Posts = filepath.Join(p.Root, p.Public, p.Posts)
	p.Full.Users = filepath.Join(p.Root, p.Users)
	p.Full.Log = filepath.Join(p.Root, p.Log)
	p.Full.Index = filepath.Join(p.Full.Views, "index.html")
	p.Full.Version = filepath.Join(p.Full.Views, ".version")

	os.MkdirAll(p.Full.Views, os.ModePerm)
	os.MkdirAll(p.Full.Public, os.ModePerm)
}

// webhook 配置
type Config struct {
	// 引擎
	*gin.Engine
	// 测试
	Debug bool `yaml:"debug"`
	// 端口 0~65535
	Port uint16 `yaml:"port"`
	// 动态
	Oid string `yaml:"oid"`
	// 资源文件夹路径
	Path Path `yaml:"path"`
	// Github 主页
	Github Github `yaml:"github"`
	// 管理员
	Administrators []string `yaml:"administrators"`
}

// 自动填充
func (r *Config) Init() {
	r.Path.Init()
	r.Github.Init()
	SetZero(&r.Port, 9000)
	SetZero(&r.Oid, "643451139714449427")
	SetNil[gin.Engine](&r.Engine, gin.Default())
	SetNil[string](&r.Administrators, []string{})
}

// 更新主页
func (r *Config) UpdateIndex() (err error) {
	// 先获取最新版本
	os.MkdirAll(r.Path.Full.Views, os.ModePerm)
	err = r.Github.GetLatestCommit()
	if err != nil {
		return
	}
	ver := resource.File{Name: r.Path.Full.Version}
	if utils.FileExist(r.Path.Full.Index) {
		s, err := ver.Read()
		if err == nil && s == r.Github.Commit.Sha {
			return nil
		}
	}
	// 再决定要不要克隆
	os.RemoveAll(r.Path.Full.Views)
	err = r.Github.Clone(r.Path.Full.Views)
	if err != nil {
		return
	}
	return ver.Write(r.Github.Commit.Sha)
}

var global *Config

func Set(c *Config) {
	if c == nil {
		c = &Config{}
		b, err := os.ReadFile("config.yml")
		if err == nil {
			yaml.Unmarshal(b, &c)
		}
	}
	c.Init()
	global = c
}

func Get() *Config {
	return global
}
