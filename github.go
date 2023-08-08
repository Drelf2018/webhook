package webhook

import (
	"os"

	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/exp/slices"
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
		Sha []byte `json:"sha"`
	} `json:"commit"`

	Path     string
	HTML     string
	Version  string
	api, git string
}

func (g *Github) init() string {
	Default(&g.Username, "Drelf2018")
	Default(&g.Repository, "gin.nana7mi.link")
	Default(&g.Branche, "gh-pages")

	g.api = API + g.Username + "/" + g.Repository + "/branches/" + g.Branche
	g.git = GIT + g.Username + "/" + g.Repository + ".git"

	return g.Repository
}

// 获取最新提交
func (g *Github) GetLatestCommit() ([]byte, error) {
	err := request.Get(g.api).Json(g)
	return g.Commit.Sha, err
}

// 克隆到文件夹
func (g Github) Clone(folder string) error {
	_, err := git.PlainClone(folder, false, &git.CloneOptions{
		URL:           g.git,
		ReferenceName: plumbing.ReferenceName(g.Branche),
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	return err
}

// 先判断文件夹存不存在 再判断主页存不存在 再判断版本对不对
func (g Github) AllExists(sha []byte) bool {
	if utils.FileNotExists(g.Path) {
		return false
	}
	if utils.FileNotExists(g.HTML) {
		return false
	}
	if utils.FileNotExists(g.Version) {
		return false
	}
	b, err := os.ReadFile(g.Version)
	if err != nil {
		return false
	}
	return slices.Equal(b, sha)
}

// 写入版本
func (g Github) Write(sha []byte) error {
	return os.WriteFile(g.Version, sha, os.ModePerm)
}
