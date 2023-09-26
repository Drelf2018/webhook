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
	"github.com/gin-gonic/gin"
)

func IsAdministrator(c *gin.Context) {
	user := GetUser(c)
	if !slices.Contains(configs.Get().Administrators, user.Uid) {
		Failed(c, 1, "您没有管理员权限")
	}
}

func runCmd(c *gin.Context, s string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", s)
	case "linux":
		cmd = exec.Command("/bin/sh", "-c", s)
	}
	cmd.Dir = configs.Get().Path.Root
	b, err := cmd.Output()
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c, CutString(b))
}

func Cmd(c *gin.Context) {
	runCmd(c, c.Param("cmd")[1:])
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
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c, u)
}

func UpdatePermission(c *gin.Context) {
	uid, permission := c.Param("uid"), c.Param("permission")
	p, err := strconv.ParseFloat(permission, 64)
	if err != nil {
		Failed(c, 1, err.Error(), "received", permission)
		return
	}
	err = user.User{Uid: uid, Permission: p}.UpdatePermission()
	if err != nil {
		Failed(c, 2, err.Error(), "received", uid)
		return
	}
	Succeed(c)
}

func Close(c *gin.Context) {
	os.Exit(0)
}

func CheckFiles(c *gin.Context) {
	err := data.CheckFiles()
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c)
}
