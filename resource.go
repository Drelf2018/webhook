package webhook

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

// 资源文件夹
type Resource struct {
	// 文件夹路径
	Path string
	// 公开子文件夹
	Public struct {
		// 子文件路径
		Path string
		// posts 数据库文件名
		Posts string
		// Github 主页
		Github Github
	}
	// users 数据库名
	Users string
}

func (r *Resource) init() {
	Default(&r.Path, "resource")
	Default(&r.Public.Path, "public")
	Default(&r.Public.Posts, "posts.db")
	Default(&r.Users, "users.db")
	repo := r.Public.Github.init()
	Default(&r.Public.Github.Path, r.ToPublic(repo))
	Default(&r.Public.Github.HTML, r.ToPublic(repo, "index.html"))
	Default(&r.Public.Github.Version, r.ToPublic(repo, ".version"))
}

func (r Resource) ToRoot(files ...string) string {
	// return filepath.Join(r.Path, files...)
	// 甚至不允许这样 这语言是有够傻逼的
	return filepath.Join(slices.Insert(files, 0, r.Path)...)
}

func (r Resource) ToPublic(files ...string) string {
	return filepath.Join(slices.Insert(files, 0, r.Path, r.Public.Path)...)
}

func (r Resource) ToIndex() string {
	return r.Public.Github.Path
}

func (r Resource) ToPostsDB() gorm.Dialector {
	return sqlite.Open(r.ToPublic(r.Public.Posts))
}

func (r Resource) ToUsersDB() gorm.Dialector {
	return sqlite.Open(r.ToRoot(r.Users))
}

func (r Resource) List() any {
	data := make(map[string]any)
	filepath.Walk(r.Path, func(path string, info fs.FileInfo, err error) error {
		temp := data
		dir, file := filepath.Split(path)
		for _, p := range strings.Split(dir, string(os.PathSeparator)) {
			if p == "" {
				continue
			}
			if temp[p] == nil {
				temp[p] = make(map[string]any)
			}
			temp = temp[p].(map[string]any)
		}
		if info.IsDir() {
			temp[file] = make(map[string]any)
		} else {
			size := float64(info.Size())
			unit := "TB"
			for _, u := range []string{"Byte", "KB", "MB", "GB", "TB"} {
				if size < 1024 {
					unit = u
					break
				}
				size /= 1024
			}
			if unit == "Byte" {
				temp[file] = fmt.Sprintf("%.0f %s", size, unit)
			} else {
				temp[file] = fmt.Sprintf("%.1f %s", size, unit)
			}
		}
		return nil
	})
	return data
}

// 主页更新
func (r Resource) IndexUpdate() (err error) {
	// 先获取最新版本
	sha, err := r.Public.Github.GetLatestCommit()
	if err != nil {
		return
	}
	if r.Public.Github.AllExists(sha) {
		return nil
	}
	// 再决定要不要克隆
	folder := r.ToIndex()
	os.RemoveAll(folder)
	err = r.Public.Github.Clone(folder)
	if err != nil {
		return
	}
	return r.Public.Github.Write(sha)
}
