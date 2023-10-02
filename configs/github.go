package configs

import (
	"fmt"
	"os"

	"github.com/Drelf2018/request"
	"github.com/Drelf2020/utils"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/exp/slices"
)

type Github struct {
	Username   string `yaml:"username" default:"Drelf2018"`
	Repository string `yaml:"repository" default:"gin.nana7mi.link"`
	Branche    string `yaml:"branche" default:"gh-pages"`

	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit" yaml:"-"`
}

func (g *Github) API() string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/branches/%v", g.Username, g.Repository, g.Branche)
}

func (g *Github) GIT() string {
	return fmt.Sprintf("https://github.com/%v/%v.git", g.Username, g.Repository)
}

func (g *Github) Sha() []byte {
	return []byte(g.Commit.Sha)
}

// 获取最新提交
func (g *Github) GetLatestCommit() error {
	return request.Get(g.API()).Json(g)
}

// 克隆到文件夹
func (g *Github) Clone(folder string) error {
	_, err := git.PlainClone(folder, false, &git.CloneOptions{
		URL:           g.GIT(),
		ReferenceName: plumbing.ReferenceName(g.Branche),
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	return err
}

// 更新主页
func (g *Github) UpdateIndex(views, index, version string) error {
	// 先获取最新版本
	if err := g.GetLatestCommit(); err != nil {
		return err
	}
	// 判断当前版本是否最新
	if utils.FileExist(index) {
		b, err := os.ReadFile(version)
		if err == nil && slices.Equal(b, g.Sha()) {
			return nil
		}
	}
	// 再决定要不要克隆
	os.RemoveAll(views)
	if err := g.Clone(views); err != nil {
		return err
	}
	return os.WriteFile(version, g.Sha(), os.ModePerm)
}
