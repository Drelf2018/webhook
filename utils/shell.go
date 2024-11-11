package utils

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/axgle/mahonia"
)

func StrToAny(conds []string) []any {
	r := make([]any, 0, len(conds))
	for _, cond := range conds {
		r = append(r, cond)
	}
	return r
}

func SplitLines(s string) (r []string) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			r = append(r, line)
		}
	}
	return
}

var enc = mahonia.NewDecoder("gbk")

func Shell(s string, dir string) ([]string, error) {
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
