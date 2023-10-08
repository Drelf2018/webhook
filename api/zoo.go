package api

import (
	"strings"

	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

// 返回成功数据
func Succeed(c *gin.Context, data ...any) {
	obj := gin.H{"code": 0}
	if len(data) == 1 {
		obj["data"] = data[0]
	} else if len(data) > 1 {
		temp := gin.H{}
		for i := 0; i < len(data); i += 2 {
			temp[data[i].(string)] = data[i+1]
		}
		obj["data"] = temp
	}
	c.JSON(200, obj)
}

// 返回错误信息
func Failed(c *gin.Context, code int, message string, data ...any) {
	obj := gin.H{"code": code, "message": message}
	for i := 0; i < len(data); i += 2 {
		obj[data[i].(string)] = data[i+1]
	}
	c.AbortWithStatusJSON(200, obj)
}

// 根据是否有错误判断返回
func Error(c *gin.Context, code int, err error, failed ...any) (hasError bool) {
	if err != nil {
		if failed == nil {
			failed = make([]any, 0)
		}
		Failed(c, code, err.Error(), failed...)
		return true
	}
	return false
}

// 根据是否有错误判断返回
func Final(c *gin.Context, code int, err error, failed []any, succeed ...any) {
	if !Error(c, code, err, failed...) {
		Succeed(c, succeed...)
	}
}

// 读取用户
func GetUser(c *gin.Context) *user.User {
	return c.MustGet("user").(*user.User)
}

// 换行分割字符串
func CutString(s []byte) []string {
	return utils.Filter(
		strings.Split(string(s), "\n"),
		func(s string) bool { return s != "" },
	)
}
