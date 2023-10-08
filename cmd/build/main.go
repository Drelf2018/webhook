//go:generate goversioninfo -icon=icon.ico
package main

import "github.com/Drelf2018/webhook"

func main() {
	webhook.Run(nil)
}
