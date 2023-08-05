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
	DIY func(r *Config)
	// 鉴权前
	BeforeAuthorize func(r *Config)
	// 鉴权后
	AfterAuthorize func(r *Config)
	// 主页 git 链接 只需填写前三项
	Github network.Github
	// 其他参数
	Map gin.H
	// 引擎
	*gin.Engine
}

// 自动填充
func (r *Config) AutoFill() {
	utils.Default(&r.Resource, "resource")
	utils.Default(&r.File, "posts.db")
	utils.Default(&r.Url, "0.0.0.0")
	utils.Default(&r.Port, "9000")
	utils.Default(&r.Github, network.Github{
		Username:   "Drelf2018",
		Repository: "nana7mi.link",
		Branche:    "gh-pages",
	})
	if r.Engine == nil {
		r.Engine = gin.Default()
	}
	if r.BeforeAuthorize == nil {
		r.BeforeAuthorize = BeforeAuthorize
	}
	if r.AfterAuthorize == nil {
		r.AfterAuthorize = AfterAuthorize
	}
	// 设置运行模式
	gin.SetMode(utils.Ternary(r.Debug, gin.DebugMode, gin.ReleaseMode))
}
