package api

import (
	"net/http"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
)

const MagicUIDKey string = "_magic_uid_key_"

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
	return ctx.GetString(MagicUIDKey)
}

func IsAdmin(ctx *gin.Context) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -1, Error: err.Error()})
	}
	user := &model.User{UID: uid}
	if err = UserDB.First(user).Error; err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -2, Error: err.Error()})
	} else if !user.Role.IsAdmin() {
		ctx.Error(ErrPermDenied)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -3, Error: ErrPermDenied.Error()})
	}
}

func IsOwner(ctx *gin.Context) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -1, Error: err.Error()})
	} else if uid != config.Role.Owner {
		ctx.Error(ErrPermDenied)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, group.Response{Code: -2, Error: ErrPermDenied.Error()})
	}
}
