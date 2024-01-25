package main

import (
	"github.com/Drelf2018/webhook"
	"github.com/gin-gonic/gin"
)

func main() {
	webhook.Default(&webhook.Config{
		Mode:  gin.DebugMode,
		Host:  "localhost",
		Port:  9000,
		Owner: "188888131",
	})
}
