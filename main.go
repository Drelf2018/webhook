package webhook

import (
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/interfaces"
	"github.com/gin-gonic/gin"
)

func init() {
	// flag.StringVar(&config.ConfigPath, "config", "./config.yml", "path to config file")
	// flag.Parse()
}

type (
	Config     = config.Config
	Github     = config.Github
	Server     = config.Server
	Permission = config.Permission
)

func SetConfig(c *config.Config) error {
	err := config.Set(c)
	if err != nil {
		return err
	}

	err = interfaces.Initial(config.Global)
	if err != nil {
		return err
	}

	gin.SetMode(config.Global.Server.Mode)
	return nil
}

func New(r *gin.Engine, c *config.Config) error {
	err := SetConfig(c)
	if err != nil {
		return err
	}
	return r.Run(config.Global.Server.Addr())
}

func Default(c *config.Config) error {
	err := SetConfig(c)
	if err != nil {
		return err
	}

	r, err := api.GetEngine()
	if err != nil {
		return err
	}
	return r.Run(config.Global.Server.Addr())
}
