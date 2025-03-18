//go:generate go test
//go:generate goversioninfo -icon=icon.ico
package main

import (
	"bufio"
	"errors"
	"fmt"
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

func pause() {
	println("按任意键继续. . .")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func log(err error) {
	os.WriteFile("error.log", []byte(fmt.Sprintf("%#v: %s", err, err)), os.ModePerm)
	fmt.Printf("发生错误: %s，", err.Error())
	pause()
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfg := &api.Config{
		Filename: "config.toml",
		Extra:    map[string]any{"password": ""},
	}

	if _, err := os.Stat(cfg.Filename); err != nil {
		err = initial.Initial(cfg)
		if err != nil {
			log(err)
			return
		}

		err = cfg.Export()
		if err != nil {
			log(err)
			return
		}

		fmt.Printf("请完善 %s 配置文件，", cfg.Filename)
		pause()
		return
	}

	err := api.Initial(cfg)
	if err != nil {
		log(err)
		return
	}

	r := &Registrar{UID: cfg.Role.Owner}
	r.Password, _ = api.LoadOrStore(cfg.Extra, "password", "")
	if r.UID == "" || r.Password == "" {
		log(ErrMissing)
		return
	}
	registrar.SetRegistrar(r)

	engine := gin.New()
	api.API.Bind(engine)

	err = webhook.Run(cfg.Server.Addr(), engine)
	if err != nil {
		log(err)
	}
}
