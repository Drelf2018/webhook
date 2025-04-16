package api

import (
	"sync"

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

// JWT 校验
func JWTAuth(ctx *gin.Context) (uid string, err error) {
	// 优先从请求头获取
	token := ctx.GetHeader("Authorization")
	// 再从请求参数中获取 用于覆盖 Cookies
	if token == "" {
		query := ctx.Request.URL.Query()
		token = query.Get("auth")
		if token != "" {
			// 避免把鉴权码写进日志
			query.Del("auth")
			ctx.Request.URL.RawQuery = query.Encode()
		}
	}
	// 最后在 Cookies 里找
	if token == "" {
		token, _ = ctx.Cookie("auth")
	}
	if token == "" {
		return "", ErrAuthNotExist
	}
	user := &UserClaims{}
	_, err = jwt.ParseWithClaims(token, user, JWTSecretKeyFn)
	if err != nil {
		return "", err
	}
	// 鉴权成功则写入 Cookies
	ctx.SetCookie("auth", token, 0, "", "", false, false)
	return user.UID, nil
}
