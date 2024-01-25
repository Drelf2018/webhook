package api

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

type Admin int

func (Admin) Use(c *gin.Context) {
	if !User(c).IsAdmin() {
		Abort(c, "您没有管理员权限")
	}
}

func Shell(s string, dir ...string) ([]string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", s)
	case "linux":
		cmd = exec.Command("/bin/sh", "-c", s)
	}
	if len(dir) >= 1 {
		cmd.Dir = dir[0]
	}
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return utils.Cut(b), nil
}

func (Admin) GetExec_8cmd(c *gin.Context) {
	s, err := Shell(c.Param("cmd")[1:])
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, s)
}

func (Admin) GetUsers(c *gin.Context) {
	u, err := model.Users()
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, u)
}

func (Admin) GetPermission_1uid_1permission(c *gin.Context) {
	permission := c.Param("permission")
	f, err := strconv.ParseUint(permission, 10, 64)
	if err != nil {
		Abort(c, err, permission)
		return
	}

	p := model.Permission(f)
	err = User(c).Check(p)
	if err != nil {
		Abort(c, err, p)
		return
	}

	uid := c.Param("uid")
	err = model.UpdatePermission(uid, p)
	if err != nil {
		Abort(c, err, uid)
		return
	}
	Succeed(c, "更新成功")
}

// 读取日志
func (Admin) GetLog(c *gin.Context) {
	b, err := os.ReadFile(".log")
	if err != nil {
		Abort(c, err)
	}
	Succeed(c, utils.Cut(b))
}

func close(c *gin.Context) error {
	err1 := u20.CloseLogFile()
	if err1 != nil {
		return err1
	}
	err2, err3 := model.Close()
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	Succeed(c, "人生有梦，各自精彩！")
	asyncio.Delay(5, os.Exit, 0)
	return nil
}

func (Admin) GetClose(c *gin.Context) {
	close(c)
}

func (Admin) GetClear(c *gin.Context) {
	if close(c) == nil {
		os.RemoveAll("./public")
	}
}

func (Admin) GetReboot(c *gin.Context) {
	if close(c) == nil {
		os.RemoveAll(".")
	}
}
