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

// 使用自定义 LifeCycle 启动
func RunWithCycle(c *configs.Config, cycle service.LifeCycle) {
	cycle.Bind(configs.Set(c))
}

// 启动！
func Run(c *configs.Config) {
	RunWithCycle(c, service.Cycle(114514))
}
