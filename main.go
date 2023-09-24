package webhook

import (
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service"
)

type (
	Config = configs.Config
	Github = configs.Github
	Path   = configs.Path
)

// 使用默认生命周期循环
var cycle service.LifeCycle = service.Cycle(114514)

func SetLifeCycle(c service.LifeCycle) {
	cycle = c
}

// 启动！
func Run(c *configs.Config) {
	configs.Set(c)
	cycle.Bind(configs.Get())
}
