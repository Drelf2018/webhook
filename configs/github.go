package configs

import (
	"os"

	"github.com/Drelf2018/request"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	GIT = "https://github.com/"
	API = "https://api.github.com/repos/"
)

type Github struct {
	Username   string
	Repository string
	Branche    string
	Commit     struct {
		Sha string `json:"sha"`
	} `json:"commit"`

	api string
	git string
}

func (g *Github) Init() {
	SetZero(&g.Username, "Drelf2018")
	SetZero(&g.Repository, "gin.nana7mi.link")
	SetZero(&g.Branche, "gh-pages")

	g.api = API + g.Username + "/" + g.Repository + "/branches/" + g.Branche
	g.git = GIT + g.Username + "/" + g.Repository + ".git"
}

// 获取最新提交
func (g *Github) GetLatestCommit() error {
	return request.Get(g.api).Json(g)
}

// 克隆到文件夹
func (g *Github) Clone(folder string) error {
	_, err := git.PlainClone(folder, false, &git.CloneOptions{
		URL:           g.git,
		ReferenceName: plumbing.ReferenceName(g.Branche),
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	return err
}
