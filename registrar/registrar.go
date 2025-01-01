package registrar

import (
	"github.com/Drelf2018/webhook"
	"github.com/gin-gonic/gin"
)

type Registrar interface {
	Initial(cfg *webhook.Config) error
	Register(ctx *gin.Context) (user any, data any, err error)
}

type RegistrarFunc func(ctx *gin.Context) (user any, data any, err error)

func (RegistrarFunc) Initial(*webhook.Config) error { return nil }

func (r RegistrarFunc) Register(ctx *gin.Context) (any, any, error) { return r(ctx) }

var _ Registrar = (RegistrarFunc)(nil)
