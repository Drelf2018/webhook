package utils

import (
	"os/exec"
	"runtime"

	"github.com/axgle/mahonia"
)

var enc = mahonia.NewDecoder("gbk")

func RunShell(s string, dir string) ([]string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", s)
	case "linux":
		cmd = exec.Command("/bin/sh", "-c", s)
	}
	cmd.Dir = dir
	b, err := cmd.Output()
	return SplitLines(enc.ConvertString(string(b))), err
}
