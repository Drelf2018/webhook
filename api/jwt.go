package api

import (
	"strings"
	"sync"

	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTSecretKeyFn(*jwt.Token) (any, error) {
	return JWTSecretKey, nil
}

var tokenIssuedAt sync.Map // map[string]int64

type UserClaims struct {
	UID      string `json:"uid" gorm:"primaryKey"`
	IssuedAt int64  `json:"iat"`
}

func (UserClaims) TableName() string {
	return "users"
}

func (c UserClaims) Valid() error {
	if v, found := tokenIssuedAt.Load(c.UID); found {
		if ver, ok := v.(int64); ok && c.IssuedAt == ver {
			return nil
		}
	}
	return ErrExpired
}

func (c UserClaims) Token(update bool) (string, error) {
	key, err := JWTSecretKeyFn(nil)
	if err != nil {
		return "", err
	}
	if update {
		err = UserDB.Updates(&c).Error
		if err != nil {
			return "", err
		}
		tokenIssuedAt.Store(c.UID, c.IssuedAt)
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(key)
}

func JWTAuth(ctx *gin.Context) (uid string, err error) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		var exists bool
		token, exists = ctx.GetQuery("auth")
		if exists {
			ctx.Request.URL.RawQuery = strings.ReplaceAll(ctx.Request.URL.RawQuery, "auth="+token, "")
			ctx.SetCookie("auth", token, 0, "", "", false, false)
		} else {
			token, _ = ctx.Cookie("auth")
		}
	}
	if token == "" {
		return "", ErrAuthNotExist
	}
	user := &UserClaims{}
	_, err = jwt.ParseWithClaims(token, user, JWTSecretKeyFn)
	if err != nil {
		return "", err
	}
	return user.UID, nil
}

func JWTUser(ctx *gin.Context) (user *model.User, err error) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		return nil, err
	}
	user = &model.User{UID: uid}
	err = UserDB.First(user).Error
	return
}
