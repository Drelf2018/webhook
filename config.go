package webhook

import (
	"github.com/Drelf2018/webhook/data"
	"github.com/gin-gonic/gin"
)

// webhook 配置
type Config struct {
	// 资源文件夹
	Resource string
	// 数据库文件名
	File string

	// 服务器启动 IP
	Url string
	// 启动端口
	Port string

	// gin 启动模式
	Debug bool

	// 自定义全接口
	//
	// DIY 不为 nil 时仅执行此函数 不会执行下面的鉴权前后函数
	DIY func(r *gin.Engine)
	// 鉴权前
	BeforeAuthorize func(r *gin.Engine)
	// 鉴权后
	AfterAuthorize func(r *gin.Engine)
}

// 自动填充
func (c *Config) AutoFill() {
	if c.Resource == "" {
		c.Resource = data.Resource
	}
	if c.File == "" {
		c.File = data.File
	}
	if c.Url == "" {
		c.Url = "0.0.0.0"
	}
	if c.Port == "" {
		c.Port = "9000"
	}
}

// 拼接地址
func (c Config) Addr() string {
	return c.Url + ":" + c.Port
}
