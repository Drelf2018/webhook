package api

import (
	"slices"

	"github.com/gin-gonic/gin"
)

func IsAdministrator(c *gin.Context) {
	user := GetUser(c)
	if !slices.Contains(config.Administrators, user.Uid) {
		Failed(c, 1, "您没有管理员权限")
	}
}
