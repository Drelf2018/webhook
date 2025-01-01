//go:generate go test
//go:generate goversioninfo -icon=icon.ico
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/gin-gonic/gin"
)

var ErrMissing = errors.New("webhook/cmd/webhook: the owner username or password is missing")

type Registrar struct {
	UID      string
	Password string
}

func (r *Registrar) Initial(cfg *webhook.Config) error {
	r.UID = cfg.Role.Owner
	r.Password = cfg.Extra["password"].(string)
	if r.UID == "" || r.Password == "" {
		return ErrMissing
	}
	return nil
}

func (r *Registrar) Register(ctx *gin.Context) (user any, data any, err error) {
	uid, password, err := registrar.BasicAuth(ctx)
	if err != nil {
		return nil, 10001, err
	}
	var payload struct {
		Name string `json:"name"`
	}
	err = ctx.ShouldBindJSON(&payload)
	if err != nil {
		return nil, 10002, err
	}
	if uid != r.UID || password == r.Password {
		return &model.User{UID: uid, Name: payload.Name, Password: password}, 0, nil
	}
	return nil, 10003, ErrMissing
}

var _ registrar.Registrar = (*Registrar)(nil)

func init() {
	registrar.SetRegistrar(&Registrar{})
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	cfg := &webhook.Config{
		Filename: "config.toml",
		Extra:    map[string]any{"password": ""},
	}
	err := webhook.Run(&api.OpenAPI{Engine: gin.New()}, cfg)
	if _, ok := err.(*fs.PathError); ok {
		initial.Initial(cfg)
		cfg.Export()
		fmt.Println("请完善配置文件 按下回车键退出")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	} else if err != nil {
		fmt.Printf("%v\n", err)
		os.WriteFile("error.log", []byte(err.Error()), os.ModePerm)
	}
}
