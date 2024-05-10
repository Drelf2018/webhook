package config

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/initial"
	"gopkg.in/yaml.v2"
)

const VERSION = "v0.15.0"

var ConfigPath string = "./config.yml"

// webhook 配置
type Config struct {
	// 路径
	Path Path `yaml:"path"`
	// Github 主页
	Github Github `yaml:"github"`
	// 服务器参数
	Server Server `yaml:"server"`
	// 权限组
	Permission Permission `yaml:"permission"`
	// 额外参数
	Extra map[string]string `yaml:"extra"`
}

func (c *Config) BeforeDefault() error {
	if c.Extra == nil {
		c.Extra = make(map[string]string)
	}
	return nil
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

var Global *Config

// Set sets c as the global config.
func Set(c *Config) error {
	// read config from file
	if c == nil {
		c = &Config{Extra: make(map[string]string)}
		b, err := os.ReadFile(ConfigPath)
		if err == nil {
			err = yaml.Unmarshal(b, c)
			if err != nil {
				return err
			}
		}
	}
	// set default value
	err := initial.Default(c)
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
