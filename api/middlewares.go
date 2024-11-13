package api

import (
	"errors"
	"net/http"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
)

const MagicUIDKey string = "__magic_uid_key__"

var (
	ErrNotAdmin = errors.New("webhook/api: no administrator permission")
	ErrNotOwner = errors.New("webhook/api: no owner permission")
)

func IsUser(ctx *gin.Context) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -1, Error: err.Error()})
	} else {
		ctx.Set(MagicUIDKey, uid)
	}
}

func GetUID(ctx *gin.Context) string {
	return ctx.MustGet(MagicUIDKey).(string)
}

func IsAdmin(ctx *gin.Context) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -1, Error: err.Error()})
	}
	user := &model.User{UID: uid}
	if err = UserDB().First(user).Error; err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -2, Error: err.Error()})
	} else if !user.Role.IsAdmin() {
		ctx.Error(ErrNotAdmin)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -3, Error: ErrNotAdmin.Error()})
	}
}

func IsOwner(ctx *gin.Context) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -1, Error: err.Error()})
	} else if uid != webhook.Global().Role.Owner {
		ctx.Error(ErrNotOwner)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -2, Error: ErrNotOwner.Error()})
	}
}
