package api

import (
	"os/exec"
	"strings"

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
	args := strings.Split(c.Param("cmd")[1:], "/")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = config.Resource.Path()
	err := cmd.Run()
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c)
}
