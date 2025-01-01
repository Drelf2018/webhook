package webhook

import (
	"context"
	"net/http"

	"github.com/Drelf2018/initial"
	"github.com/gin-gonic/gin"
)

var stop func() bool
var Quit, cancel = context.WithCancel(context.Background())

func Shutdown() {
	stop()
	cancel()
}

type Handler interface {
	http.Handler
	Initial(*Config) error
}

func run(handler Handler, cfg *Config) error {
	if cfg == nil {
		cfg = &Config{Filename: "config.yml"}
	}
	if cfg.Role.Admin == nil {
		cfg.Role.Admin = make([]string, 0)
	}
	if cfg.Extra == nil {
		cfg.Extra = make(map[string]any)
	}

	err := cfg.Import()
	if err != nil {
		return err
	}
	err = initial.Initial(cfg)
	if err != nil {
		return err
	}
	err = cfg.Export()
	if err != nil {
		return err
	}
	err = handler.Initial(cfg)
	if err != nil {
		return err
	}

	switch cfg.Server.Mode {
	case gin.ReleaseMode, gin.DebugMode, gin.TestMode:
		gin.SetMode(cfg.Server.Mode)
	}

	return nil
}

func Run(handler Handler, cfg *Config) error {
	err := run(handler, cfg)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: handler,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go srv.ListenAndServe()

	<-Quit.Done()

	return srv.Shutdown(context.Background())
}

func RunForever(handler Handler, cfg *Config) error {
	err := run(handler, cfg)
	if err != nil {
		return err
	}
	return http.ListenAndServe(cfg.Server.Addr(), handler)
}
