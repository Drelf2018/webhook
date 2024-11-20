//go:generate go test
//go:generate goversioninfo -icon=icon.ico
package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/gin-gonic/gin"
)

var ErrMissing = errors.New("webhook/cmd/webhook: the owner username or password is missing")

func init() {
	registrar.SetRegistrarFunc(func(ctx *gin.Context) (user any, data any, err error) {
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
		owner := webhook.Global().Role.Owner
		pwd := webhook.Global().Extra["password"].(string)
		if owner == "" || pwd == "" {
			return nil, 10003, ErrMissing
		}
		if uid != owner || pwd == password {
			return &model.User{UID: uid, Name: payload.Name, Password: password}, 0, nil
		}
		return nil, 10004, ErrMissing
	})
}

func main() {
	addr, err := webhook.Initial(&webhook.Config{
		Filename: "config.toml",
		Extra:    map[string]any{"password": ""},
	})
	if err != nil {
		os.WriteFile("error.log", []byte(err.Error()), os.ModePerm)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.New(),
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go srv.ListenAndServe()

	<-webhook.Quit.Done()

	err = srv.Shutdown(context.Background())
	if err != nil {
		os.WriteFile("error.log", []byte(err.Error()), os.ModePerm)
	}
}
