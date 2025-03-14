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

func logError(err error) {
	println(err.Error())
	os.WriteFile("error.log", []byte(err.Error()), os.ModePerm)
}

func pause() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func main() {
	cfg := &api.Config{
		Filename: "config.toml",
		Server:   api.Server{Mode: gin.ReleaseMode},
		Extra:    map[string]any{"password": ""},
	}
	engine := gin.New()
	err := api.Initial(engine, cfg)

	if _, ok := err.(*fs.PathError); ok {
		initial.Initial(cfg)
		cfg.Export()
		fmt.Println("请完善配置文件 按下回车键退出")
		pause()
		return
	} else if err != nil {
		logError(err)
		pause()
		return
	}

	r := &Registrar{
		UID:      cfg.Role.Owner,
		Password: cfg.Extra["password"].(string),
	}
	if r.UID == "" || r.Password == "" {
		logError(ErrMissing)
		pause()
		return
	}

	registrar.SetRegistrar(r)

	err = webhook.Run(cfg.Server.Addr(), engine)
	if err != nil {
		logError(err)
		pause()
	}
}
