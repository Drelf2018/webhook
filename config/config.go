package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/utils"
	"gopkg.in/yaml.v2"
)

const VERSION = "v0.15.0"

var ConfigPath string = "./config.yml"

type Path struct {
	Root   string `yaml:"root"   default:"resource"`
	Logs   string `yaml:"logs"   default:"logs"     join:"Root"`
	UserDB string `yaml:"userDB" default:"users.db" join:"Root"`
	Public string `yaml:"public" default:"public"   join:"Root"`
	BlogDB string `yaml:"blogDB" default:"blogs.db" join:"Public"`
	Backup string `yaml:"backup" default:"backup"   join:"Public"`

	full *Path
}

func (p *Path) Full() *Path {
	return p.full
}

func (p *Path) AfterInitial() (err error) {
	p.full, err = utils.NewJoin(*p)
	if err != nil {
		return
	}

	// err = os.MkdirAll(p.full.Logs, os.ModePerm)
	// if err != nil {
	// 	return err
	// }
	return nil
	// utils.SetOutputFile(p.full.Logs)

	// return dao.Open(p.Full().PostDB, p.Full().UserDB)
}

var _ initial.AfterInitial2 = (*Path)(nil)

type Server struct {
	Mode string `yaml:"mode" default:"release"`   // 模式
	Host string `yaml:"host" default:"localhost"` // 主机
	Port uint16 `yaml:"port" default:"9000"`      // 端口 0~65535
}

func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Permission struct {
	Owner          string   `yaml:"owner"`          // 所有者
	Administrators []string `yaml:"administrators"` // 管理员
}

// webhook 配置
type Config struct {
	Path       Path           `yaml:"path"`       // 各文件路径
	Server     Server         `yaml:"server"`     // 服务器参数
	Permission Permission     `yaml:"permission"` // 权限组
	Extra      map[string]any `yaml:"extra"`      // 额外项
}

func (c *Config) BeforeInitial() {
	if c.Extra == nil {
		c.Extra = make(map[string]any)
	}
}

func (c *Config) Export() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(ConfigPath), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath, b, os.ModePerm)
}

var _ initial.BeforeInitial1 = (*Config)(nil)

var Global *Config

// Set sets c as the global config.
func Set(c *Config) error {
	// read config from file
	if c == nil {
		c = &Config{Extra: make(map[string]any)}
		b, err := os.ReadFile(ConfigPath)
		if err == nil {
			err = yaml.Unmarshal(b, c)
			if err != nil {
				return err
			}
		}
	}
	// set default value
	err := initial.Initial(c)
	if err != nil {
		return err
	}
	// write config to file
	err = c.Export()
	if err != nil {
		return err
	}
	// update global config
	Global = c
	return nil
}
