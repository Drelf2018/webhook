package webhook

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 设置默认值
func Default[T comparable](a *T, b T) {
	var zero T
	if *a == zero {
		*a = b
	}
}

// 网络
type Network struct {
	// 服务器启动 IP
	Url string
	// 启动端口
	Port int
}

func (n *Network) init() {
	Default(&n.Url, "0.0.0.0")
	Default(&n.Port, 9000)
}

func (n Network) Addr() string {
	return n.Url + ":" + strconv.Itoa(n.Port)
}

// webhook 配置
type Config struct {
	// 资源
	Resource
	// 网络
	Network
	// 生命周期
	LifeCycle
	// 引擎
	*gin.Engine
}

// 自动填充
func (r *Config) init() {
	r.Resource.init()
	r.Network.init()
	if r.LifeCycle == nil {
		r.LifeCycle = Cycle{}
	}
	if r.Engine == nil {
		r.Engine = gin.Default()
	}
}
