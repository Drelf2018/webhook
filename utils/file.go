package utils

import "os"

func FileNotExists(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}
