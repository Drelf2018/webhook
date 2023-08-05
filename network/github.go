package network

import (
	"os"
	"path"

	"github.com/Drelf2018/webhook/utils"
	"github.com/Drelf2020/utils/request"
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
}

// 仓库对应 api
func (g Github) ToAPI() string {
	return API + g.Username + "/" + g.Repository + "/branches/" + g.Branche
}

// 转 git
func (g Github) ToGIT() string {
	return GIT + g.Username + "/" + g.Repository + ".git"
}

// 最后一次提交记录
func (g Github) ToData() []byte {
	return []byte(g.Commit.Sha)
}

// 版本文件路径
func (g Github) Version() string {
	return path.Join(g.Repository, ".version")
}

// 判断版本文件是否存在
func (g Github) NoVersion() bool {
	return utils.FileNotExists(g.Version())
}

// 获取最新版本
func (g *Github) GetLastCommit() error {
	if g.Commit.Sha != "" {
		return nil
	}
	return request.Get(g.ToAPI()).Json(g)
}

// 写入版本
func (g Github) Write() bool {
	return os.WriteFile(g.Version(), g.ToData(), os.ModePerm) == nil
}

// 检查版本是否正确
func (g Github) Check() bool {
	b, err := os.ReadFile(g.Version())
	if err != nil {
		return false
	}
	return string(b) == g.Commit.Sha
}
