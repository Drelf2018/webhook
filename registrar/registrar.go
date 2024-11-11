package registrar

import (
	"github.com/gin-gonic/gin"
)

type Registrar interface {
	Initial(extra map[string]any) error
	Register(ctx *gin.Context) (user any, data any, err error)
}

type RegistrarFunc func(ctx *gin.Context) (user any, data any, err error)

func (RegistrarFunc) Initial(map[string]any) error { return nil }

func (r RegistrarFunc) Register(ctx *gin.Context) (any, any, error) { return r(ctx) }

var _ Registrar = (RegistrarFunc)(nil)
