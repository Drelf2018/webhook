package main

import (
	"os"
	"os/exec"

	"github.com/Drelf2018/webhook"
)

func main() {
	_, err := os.Stat("nana7mi.link")
	if err != nil {
		exec.Command("git", "clone", "-b", "gh-pages", "https://github.com/Drelf2018/nana7mi.link.git").Run()
	}
	webhook.Run(&webhook.Config{})
}
