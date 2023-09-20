package api

import (
	"os/exec"

	"golang.org/x/exp/slices"

	"github.com/gin-gonic/gin"
)

func IsAdministrator(c *gin.Context) {
	user := GetUser(c)
	if !slices.Contains(config.Administrators, user.Uid) {
		Failed(c, 1, "您没有管理员权限")
	}
}

func Cmd(c *gin.Context) {
	cmd := exec.Command("/bin/sh", "-c", c.Param("cmd")[1:])
	cmd.Dir = config.Resource.Path()
	b, err := cmd.Output()
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c, string(b))
}
