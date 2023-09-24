package main

import "github.com/Drelf2018/webhook"

func main() {
	webhook.Run(&webhook.Config{Debug: true, Administrators: []string{"188888131"}})
}
