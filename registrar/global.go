package registrar

import (
	"github.com/Drelf2018/webhook"
	"github.com/gin-gonic/gin"
)

var registrar Registrar
var initialized bool

func SetRegistrar(reg Registrar) {
	registrar = reg
	initialized = false
}

func SetRegistrarFunc(fn RegistrarFunc) {
	registrar = fn
	initialized = true
}

func Register(ctx *gin.Context) (user any, data any, err error) {
	if registrar == nil {
		return nil, -1, ErrNoRegistrar
	}
	if !initialized {
		cfg := webhook.Global()
		err = registrar.Initial(cfg.Extra)
		if err != nil {
			return nil, -2, err
		}
		err = cfg.Export()
		if err != nil {
			return nil, -3, err
		}
		initialized = true
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
