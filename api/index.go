package api

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var indexFiles map[string]string

func init() {
	indexFiles = make(map[string]string)
}

func Index(ctx *gin.Context) {
	if file, ok := indexFiles[ctx.Request.URL.Path]; ok {
		if redirect := ctx.Query("redirect"); redirect != "" {
			if file, ok = indexFiles["/"+redirect]; ok {
				ctx.File(file)
				ctx.Abort()
			}
		} else {
			ctx.File(file)
			ctx.Abort()
		}
	}
}

func LoadFile(root, file string) {
	repl := strings.NewReplacer(root, "", "\\", "/")
	filename := repl.Replace(file)
	indexFiles[filename] = file
	if strings.HasPrefix(filename, "/index") && strings.HasSuffix(filename, ".html") {
		indexFiles["/"] = file
		indexFiles["/index.html"] = file
		version.Index = append(version.Index, filename)
	}
}

func LoadDir(root, path string) error {
	path = filepath.Clean(path)
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		file := filepath.Join(path, name)
		if entry.IsDir() {
			err := LoadDir(root, file)
			if err != nil {
				return err
			}
		} else {
			LoadFile(root, file)
		}
	}
	return nil
}
