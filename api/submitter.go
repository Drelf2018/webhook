package api

import (
	"strings"

	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/model"

	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

var log = u20.SetTimestampFormat("2006-01-02 15:04:05")

func init() {
	u20.SetOutputFile(".log")
}

type Submitter int

func (Submitter) Use(c *gin.Context) {
	// 从 headers 或者 query 获取身份码
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.GetQuery("auth")
	}
	if token == "" {
		token, _ = c.GetQuery("Auth")
	}
	if token == "" {
		token, _ = c.GetQuery("authorization")
	}
	if token == "" {
		token, _ = c.GetQuery("Authorization")
	}
	if token == "" {
		Abort(c, "需要身份鉴权")
		return
	}

	u := model.QueryUser(token)
	if u == nil {
		Abort(c, "鉴权失败", token)
		return
	}

	// 修改在线时间戳
	utils.Timer(u.Uid)
	c.Set("user", u)

	// 打印日志 不记录 ping 请求
	if !strings.Contains(c.Request.URL.Path, "/ping") {
		log.Infof("%v %v \"%v\"", u, c.Request.Method, c.Request.URL)
	}
}

// 更新在线时间
func (Submitter) GetPing(c *gin.Context) {
	utils.Timer(User(c).Uid)
}

// 获取自身信息
func (Submitter) GetMe(c *gin.Context) {
	Succeed(c, c.MustGet("user"))
}

// 主动更新主页
func (Submitter) GetUpdate(c *gin.Context) {
	cfg := config.Global
	err := cfg.UpdateIndex()
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, cfg.Github.Commit.Sha)
}

// 更新监听列表
func (Submitter) GetModify(c *gin.Context) {
	u := User(c)
	u.Follow = c.QueryArray("follow")
	err := u.Update()
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, u.Follow)
}

// 新增任务
func (Submitter) PostAdd(c *gin.Context) {
	job := model.Job{}
	err := c.Bind(&job)
	if err != nil {
		Abort(c, err)
		return
	}
	u := User(c)
	u.Jobs = append(u.Jobs, job)
	err = u.Update()
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, u.Jobs)
}

// 移除任务
func (Submitter) GetRemove(c *gin.Context) {
	u := User(c)
	err := u.RemoveJobs(c.QueryArray("jobs"))
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, u.Jobs)
}

// 测试单个任务
func (Submitter) PostTest(c *gin.Context) {
	job := make([]model.Job, 1)
	err := c.Bind(&job[0])
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, model.TestPost.Send(job)[0])
}

// 测试任务
func (Submitter) GetTests(c *gin.Context) {
	jobs := model.GetJobsByID(User(c).Uid, c.QueryArray("jobs"))
	Succeed(c, model.TestPost.Send(jobs))
}

var ErrNoRepost = `Form key "repost" not found. A "null" value must be pass in if there is no repost.`

// 提交博文
func (Submitter) PostSubmit(c *gin.Context) {
	// 检验数据合法
	v, ok := c.GetPostForm("repost")
	if !ok {
		Abort(c, ErrNoRepost, v)
		return
	}
	// 初始化 post
	post := model.Post{Submitter: User(c)}
	// 绑定其他数据
	err := c.Bind(&post)
	if err != nil {
		Abort(c, err)
		return
	}
	// 不允许提交已储存的博文刷积分
	if post.Exists() {
		Abort(c, "该博文已被提交过")
		return
	}
	// 检查该用户是否已提交过
	err = post.Submit()
	if err != nil {
		Abort(c, err)
	}
	Succeed(c, "提交成功")
}
