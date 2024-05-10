package main

import (
	"github.com/Drelf2018/webhook"
	_ "github.com/Drelf2018/webhook/registrar"
	"github.com/gin-gonic/gin"
)

func main() {
	err := webhook.Default(&webhook.Config{
		Server: webhook.Server{
			Mode: gin.DebugMode,
			Host: "localhost",
			Port: 9000,
		},
		Permission: webhook.Permission{
			Owner: "188888131",
		},
	})
	if err != nil {
		panic(err)
	}
}
