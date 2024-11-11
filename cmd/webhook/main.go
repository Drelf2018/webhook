//go:generate go test
//go:generate goversioninfo -icon=icon.ico
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/api"
)

func main() {
	addr, err := webhook.Initial(&webhook.Config{Filename: "config.toml"})
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
