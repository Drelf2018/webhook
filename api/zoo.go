package api

import (
	"os"

	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

var config *configs.Config

func SetConfig(c *configs.Config) *configs.Config {
	if c == nil {
		c = &configs.Config{}
		b, err := os.ReadFile("config.yml")
		if err == nil {
			yaml.Unmarshal(b, &c)
		}
	}
	config = c.Init()
	return config
}

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

// 读取用户
func GetUser(c *gin.Context) *user.User {
	return c.MustGet("user").(*user.User)
}
