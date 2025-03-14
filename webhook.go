package webhook

import (
	"context"
	"net/http"

	_ "unsafe"
)

//go:linkname running
var running context.Context

//go:linkname cancel
var cancel context.CancelFunc

func init() {
	running, cancel = context.WithCancel(context.Background())
}

func Run(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go srv.ListenAndServe()

	<-running.Done()

	return srv.Shutdown(context.Background())
}
