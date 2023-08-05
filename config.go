package webhook

import (
	"github.com/Drelf2018/webhook/network"
	"github.com/Drelf2018/webhook/utils"
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
	// 主页 git 链接 只需填写前三项
	Github network.Github
	// 其他参数
	Map gin.H
}

// 自动填充
func (c *Config) AutoFill() {
	utils.Default(&c.Resource, "resource")
	utils.Default(&c.File, "posts.db")
	utils.Default(&c.Url, "0.0.0.0")
	utils.Default(&c.Port, "9000")
	utils.Default(&c.Github, network.Github{
		Username:   "Drelf2018",
		Repository: "nana7mi.link",
		Branche:    "gh-pages",
	})
}

// 拼接地址
func (c Config) Addr() string {
	return c.Url + ":" + c.Port
}
