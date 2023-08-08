package webhook

import (
	"github.com/gin-gonic/gin"
)

// 设置默认值
func Default[T comparable](a *T, b T) {
	var zero T
	if *a == zero {
		*a = b
	}
}

// webhook 配置
type Config struct {
	// 端口 0~65535
	Port uint16
	// 资源
	Resource
	// 生命周期
	LifeCycle
	// 引擎
	*gin.Engine
	// 管理员
	Administrators []string
}

// 自动填充
func (r *Config) init() {
	Default(&r.Port, 9000)
	r.Resource.init()
	if r.LifeCycle == nil {
		r.LifeCycle = Cycle(0)
	}
	if r.Engine == nil {
		r.Engine = gin.Default()
	}
}
