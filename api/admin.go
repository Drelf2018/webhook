package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

func IsAdministrator(c *gin.Context) {
	user := GetUser(c)
	if !slices.Contains(configs.Get().Administrators, user.Uid) {
		Failed(c, 1, "您没有管理员权限")
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
	} else {
		cmd.Dir = configs.Get().Path.Root
	}
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return CutString(b), nil
}

func Cmd(c *gin.Context) {
	s, err := Shell(c.Param("cmd")[1:])
	Final(c, 1, err, nil, s)
}

type users []user.User

func (u users) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBufferString("[")
	for i, l := 0, len(u); i < l; i++ {
		j, _ := json.Marshal(u[i].Jobs)
		k, _ := u[i].Listening.Value()
		buf.WriteString(fmt.Sprintf(
			`{"uid":"%v","token":"%v","permission":"%v","jobs":%v,"listening":[%v]}`,
			u[i].Uid, u[i].Token, u[i].Permission, string(j), k,
		))
	}
	buf.WriteByte(']')
	return json.RawMessage(buf.Bytes()), nil
}

func Users(c *gin.Context) {
	var u users
	err := user.Users.Preloads(&u).Error()
	Final(c, 1, err, nil, u)
}

func UpdatePermission(c *gin.Context) {
	uid, permission := c.Param("uid"), c.Param("permission")
	p, err := strconv.ParseFloat(permission, 64)
	if Error(c, 1, err, "received", permission) {
		return
	}
	err = user.User{Uid: uid, Permission: p}.UpdatePermission()
	Final(c, 2, err, []any{"received", uid})
}

func Close(c *gin.Context) {
	u20.CloseLogFile()
	data.Data.Close()
	user.Users.Close()
	utils.Delay(5, os.Exit, 0)
	Succeed(c, "人生有梦，各自精彩！")
}

func Clear(c *gin.Context) {
	Close(c)
	os.RemoveAll(configs.Get().Path.Full.Public)
}

func Reboot(c *gin.Context) {
	Close(c)
	os.RemoveAll(configs.Get().Path.Root)
}

func CheckFiles(c *gin.Context) {
	Final(c, 1, data.CheckFiles(), nil, "检查中")
}
