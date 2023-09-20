package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/webhook/service/user"
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
	err := user.Preloads(&u)
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
