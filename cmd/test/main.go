package main

import "github.com/Drelf2018/webhook"

func main() {
	webhook.Debug(&webhook.Config{Port: 9000, Administrators: []string{"188888131"}})
}
