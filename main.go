package webhook

import (
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service"
	"github.com/gin-gonic/gin"
)

type (
	Config = configs.Config
	Github = configs.Github
	Path   = configs.Path
)

var Default = &Config{}
var cycle configs.LifeCycle = service.Cycle(114514)

func SetLifeCycle(c configs.LifeCycle) {
	cycle = c
}

// 测试版
func Debug(c *configs.Config) {
	gin.SetMode(gin.DebugMode)
	cycle.Bind(api.SetConfig(c))
}

// 发行版
func Release(c *configs.Config) {
	gin.SetMode(gin.ReleaseMode)
	cycle.Bind(api.SetConfig(c))
}
