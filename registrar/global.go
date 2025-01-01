package registrar

import (
	"github.com/Drelf2018/webhook"
	"github.com/gin-gonic/gin"
)

var registrar Registrar

func SetRegistrar(reg Registrar) {
	registrar = reg
}

func SetRegistrarFunc(fn RegistrarFunc) {
	registrar = fn
}

func Initial(cfg *webhook.Config) error {
	if registrar == nil {
		return ErrNoRegistrar
	}
	return registrar.Initial(cfg)
}

func Register(ctx *gin.Context) (user any, data any, err error) {
	if registrar == nil {
		return nil, -1, ErrNoRegistrar
	}
	return registrar.Register(ctx)
}

func BasicAuth(ctx *gin.Context) (uid, password string, err error) {
	uid, password, ok := ctx.Request.BasicAuth()
	switch {
	case !ok:
		err = ErrNoAuth
	case uid == "":
		err = ErrNoUID
	case password == "":
		err = ErrNoPassword
	}
	return
}
