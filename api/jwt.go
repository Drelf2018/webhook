package api

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var JWTSecretKeyFn = func(*jwt.Token) (any, error) {
	return JWTSecretKey, nil
}

var tokenIssuedAt sync.Map // map[string]int64

// 用户声明
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

func (c UserClaims) Token() (string, error) {
	key, err := JWTSecretKeyFn(nil)
	if err != nil {
		return "", err
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(key)
}

// JWT 校验
func JWTAuth(ctx *gin.Context) (uid string, err error) {
	// 优先从请求头获取
	token := ctx.GetHeader("Authorization")
	// 再从请求参数中获取
	if token == "" {
		query := ctx.Request.URL.Query()
		token = query.Get("auth")
		if token != "" {
			// 避免把鉴权码写进日志
			query.Del("auth")
			ctx.Request.URL.RawQuery = query.Encode()
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
