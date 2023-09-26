package configs

import (
	"fmt"
	"os"

	"github.com/Drelf2018/request"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Github struct {
	Username   string `yaml:"username"`
	Repository string `yaml:"repository"`
	Branche    string `yaml:"branche"`

	Sha    []byte `yaml:"-"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit" yaml:"-"`
}

func (g *Github) Init() {
	SetZero(&g.Username, "Drelf2018")
	SetZero(&g.Repository, "gin.nana7mi.link")
	SetZero(&g.Branche, "gh-pages")
}

func (g *Github) API() string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/branches/%v", g.Username, g.Repository, g.Branche)
}

func (g *Github) GIT() string {
	return fmt.Sprintf("https://github.com/%v/%v.git", g.Username, g.Repository)
}

// 获取最新提交
func (g *Github) GetLatestCommit() error {
	defer func() { g.Sha = []byte(g.Commit.Sha) }()
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
