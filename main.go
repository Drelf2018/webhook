package webhook

import (
	"github.com/Drelf2018/gins"
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/config"
	"github.com/gin-gonic/gin"
)

type (
	Config = config.Config
	Github = config.Github
)

func SetConfig(c *config.Config) error {
	err := config.Set(c)
	if err != nil {
		return err
	}
	gin.SetMode(config.Global.Mode)
	return nil
}

func New(r *gin.Engine, c *config.Config) error {
	SetConfig(c)
	return r.Run(config.Global.Addr())
}

func Default(c *config.Config) error {
	SetConfig(c)
	r, err := gins.Default(api.Api{})
	if err != nil {
		return err
	}
	return r.Run(config.Global.Addr())
}
