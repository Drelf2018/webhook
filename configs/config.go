package configs

import (
	"os"

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
}

func (p *Path) Init() {
	SetZero(&p.Root, "resource")
	SetZero(&p.Views, "views")
	SetZero(&p.Public, "public")
	SetZero(&p.Posts, "posts.db")
	SetZero(&p.Users, "users.db")
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
	// 资源文件管理器
	Resource resource.Explorer
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
	if r.Resource == nil {
		r.Resource = new(resource.Resource).Init(r.Path.Root).Root
	}
}

// 更新主页
func (r *Config) UpdateIndex() (err error) {
	// 先获取最新版本
	views := r.Resource.MakeTo(r.Path.Views)
	views.MkdirAll()

	err = r.Github.GetLatestCommit()
	if err != nil {
		return
	}

	if views.Find("index.html") != nil {
		if ver := views.Find(".version"); ver != nil {
			if ver.MustRead() == r.Github.Commit.Sha {
				return nil
			}
		}
	}

	// 再决定要不要克隆
	folder := views.Path()
	views.RemoveAll()
	err = r.Github.Clone(folder)
	if err != nil {
		return
	}

	ver, _ := r.Resource.MakeTo(r.Path.Views).Touch(".version", 0)
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
