package interfaces

import (
	"errors"

	"github.com/Drelf2018/webhook/config"
	"github.com/gin-gonic/gin"
)

type Registrar interface {
	Initial(*config.Config) error
	Token(*gin.Context) (data any, err error)
	Register(*gin.Context) (uid string, err error)
}

var defaultRegistrar Registrar

func SetRegistrar(reg Registrar) {
	defaultRegistrar = reg
}

var ErrNoRegistrar = errors.New("webhook/interfaces: registrar is not set")

func Initial(c *config.Config) error {
	if defaultRegistrar == nil {
		return ErrNoRegistrar
	}
	return defaultRegistrar.Initial(c)
}

func Token(ctx *gin.Context) (data any, err error) {
	return defaultRegistrar.Token(ctx)
}

func Register(ctx *gin.Context) (uid string, err error) {
	return defaultRegistrar.Register(ctx)
}
