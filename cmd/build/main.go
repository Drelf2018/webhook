//go:generate goversioninfo -icon=icon.ico
package main

import (
	"os"

	"github.com/Drelf2018/webhook"
)

func main() {
	err := webhook.Default(nil)
	if err != nil {
		os.WriteFile("error.log", []byte(err.Error()), os.ModePerm)
	}
}
