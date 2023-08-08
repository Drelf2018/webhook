package webhook

import (
	"github.com/gin-gonic/gin"
)

func Run(r *Config) {
	// 自动填充配置
	r.init()
	// 初始化
	r.OnCreate(r)
	// 跨域设置
	r.OnCors(r)
	// 静态资源绑定
	r.OnStatic(r)
	// 无鉴权接口
	r.BeforeAuthorize(r)
	// 鉴权
	r.OnAuthorize(r)
	// 有鉴权接口
	r.AfterAuthorize(r)
	// 鉴定管理员权限
	r.OnAdmin(r)
	// 管理员接口
	r.AfterAdmin(r)
	// 运行 gin 服务器 默认 0.0.0.0:9000
	r.Run(r.Addr())
}

// 测试版
func Debug(r *Config) {
	gin.SetMode(gin.DebugMode)
	Run(r)
}

// 发行版
func Release(r *Config) {
	gin.SetMode(gin.ReleaseMode)
	Run(r)
}
