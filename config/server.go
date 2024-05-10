package config

import "fmt"

type Server struct {
	// 模式
	Mode string `default:"release"`
	// 主机
	Host string `yaml:"host"`
	// 端口 0~65535
	Port uint16 `yaml:"port" default:"9000"`
}

func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
