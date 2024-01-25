package config

import (
	"fmt"
	"os"

	"github.com/Drelf2018/initial"
	"gopkg.in/yaml.v2"
)

const VERSION = "v0.14.0"

// webhook 配置
type Config struct {
	// Github 主页
	Github `yaml:"github" default:"MustUpdate"`
	// 模式
	Mode string `default:"release"`
	// 主机
	Host string `yaml:"host"`
	// 端口 0~65535
	Port uint16 `yaml:"port" default:"9000"`
	// 动态
	Oid string `yaml:"oid" default:"643451139714449427"`
	// 所有者
	Owner string `yaml:"owner"`
	// 管理员
	Administrators []string `yaml:"administrators"`
	// 信任者
	Trustors []string `yaml:"trustors"`
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

var Global *Config

// Set sets c as the global config.
func Set(c *Config) error {
	// read config from file
	if c == nil {
		c = &Config{}
		b, err := os.ReadFile("config.yml")
		if err == nil {
			err = yaml.Unmarshal(b, &c)
			if err != nil {
				return err
			}
		}
	}
	// set default value
	initial.Default(c)
	// write config to file
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	os.WriteFile("./config.yml", b, os.ModePerm)
	// update global config
	Global = c
	return nil
}
