package configs

import (
	"os"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

// 设置引用类型对象默认值
func SetNil[C any, T utils.CanNil[C]](a *T, b T) {
	if *a == nil {
		*a = b
	}
}

// 资源文件夹路径
type Path struct {
	Root    string `yaml:"root" default:"resource"`
	Views   string `yaml:"views" default:"views" abs:"Root"`
	Public  string `yaml:"public" default:"public" abs:"Root"`
	Posts   string `yaml:"posts" default:"posts.db" abs:"Public"`
	Users   string `yaml:"users" default:"users.db" abs:"Root"`
	Log     string `yaml:"log" default:".log" abs:"Root"`
	Index   string `yaml:"index" default:"index.html" abs:"Views"`
	Version string `yaml:"version" default:".version" abs:"Views"`

	Full *Path `yaml:"-" default:"initial.Abs"`
}

func (p *Path) MkdirAll(_ *Config) {
	os.MkdirAll(p.Full.Views, os.ModePerm)
	os.MkdirAll(p.Full.Public, os.ModePerm)
}

// webhook 配置
type Config struct {
	// 引擎
	*gin.Engine `yaml:"-"`
	// 测试
	Debug bool `yaml:"debug"`
	// 端口 0~65535
	Port int64 `yaml:"port" default:"9000"`
	// 动态
	Oid string `yaml:"oid" default:"643451139714449427"`
	// 资源文件夹路径
	Path Path `yaml:"path" default:"initial.Default;MkdirAll"`
	// Github 主页
	Github Github `yaml:"github" default:"initial.Default"`
	// 管理员
	Administrators []string `yaml:"administrators"`
}

// 自动填充
func (c *Config) Init() {
	// 设置模式
	gin.SetMode(utils.Ternary(c.Debug, gin.DebugMode, gin.ReleaseMode))
	if c.Engine == nil {
		c.Engine = gin.Default()
	}
	initial.Default(c)
}

// 更新主页
func (c *Config) UpdateIndex() error {
	return c.Github.UpdateIndex(c.Path.Full.Views, c.Path.Full.Index, c.Path.Full.Version)
}

var global *Config

func Set(c *Config) *Config {
	if c == nil {
		c = &Config{}
		b, err := os.ReadFile("config.yml")
		if err == nil {
			yaml.Unmarshal(b, &c)
		}
	}
	c.Init()

	b, err := yaml.Marshal(c)
	if err == nil {
		os.WriteFile("config.yml", b, os.ModePerm)
	}
	global = c
	return global
}

func Get() *Config {
	return global
}
