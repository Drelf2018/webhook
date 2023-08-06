package webhook

import (
	"os"

	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	GIT = "https://github.com/"
	API = "https://api.github.com/repos/"
)

type Github struct {
	*Resource
	Username   string
	Repository string
	Branche    string
	Commit     struct {
		Sha string `json:"sha"`
	} `json:"commit"`
}

func (g *Github) fill(res *Resource) {
	g.Resource = res
	Default(&g.Username, "Drelf2018")
	Default(&g.Repository, "nana7mi.link")
	Default(&g.Branche, "gh-pages")
}

// 仓库对应 api
func (g Github) ToAPI() string {
	return API + g.Username + "/" + g.Repository + "/branches/" + g.Branche
}

// 转 git
func (g Github) ToGIT() string {
	return GIT + g.Username + "/" + g.Repository + ".git"
}

// 转莫名其妙
func (g Github) ToReference() plumbing.ReferenceName {
	return plumbing.ReferenceName(g.Branche)
}

// 最后一次提交记录
func (g Github) ToData() []byte {
	return []byte(g.Commit.Sha)
}

// 版本文件路径
func (g Github) Version() string {
	return g.ToPublic(g.Repository, ".version")
}

// 主页路径
func (g Github) Index() string {
	return g.ToPublic(g.Repository, "index.html")
}

// 判断主页是否存在
func (g Github) NoIndex() bool {
	return utils.FileNotExists(g.Index())
}

// 获取最新版本
func (g *Github) GetLastCommit() error {
	return request.Get(g.ToAPI()).Json(g)
}

// 写入版本
func (g Github) Write() error {
	return os.WriteFile(g.Version(), g.ToData(), os.ModePerm)
}

// 检查版本是否正确
func (g Github) Check() bool {
	b, err := os.ReadFile(g.Version())
	if err != nil {
		return false
	}
	return string(b) == g.Commit.Sha
}
