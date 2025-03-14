package registrar

import (
	"github.com/gin-gonic/gin"
)

type Registrar interface {
	Register(ctx *gin.Context) (user any, data any, err error)
}

type RegistrarFunc func(ctx *gin.Context) (user any, data any, err error)

func (r RegistrarFunc) Register(ctx *gin.Context) (any, any, error) { return r(ctx) }

var _ Registrar = (RegistrarFunc)(nil)
