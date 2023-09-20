package api

import (
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

func IsSubmitter(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("Authorization")
	}
	if token == "" {
		Failed(c, 1, "你是不是调错接口了啊")
		return
	}
	user := user.Query(token)
	if user == nil {
		Failed(c, 2, "鉴权失败", "received", token)
		return
	}
	utils.Timer(user.Uid)
	c.Set("user", user)
}

func Ping(c *gin.Context) {
	utils.Timer(GetUser(c).Uid)
}

// 获取自身信息
func Me(c *gin.Context) {
	u := GetUser(c)
	Succeed(c, "uid", u.Uid, "permission", u.Permission, "listening", u.Listening, "jobs", u.Jobs)
}

// 主动更新主页
func Update(c *gin.Context) {
	err := config.UpdateIndex()
	if err != nil {
		Failed(c, 1, err.Error())
		return
	}
	Succeed(c)
}

// 更新监听列表
func ModifyListening(c *gin.Context) {
	u := GetUser(c)
	u.Listening = c.QueryArray("listen")
	err := u.Update()
	if err != nil {
		Failed(c, 1, err.Error(), "received", u.Listening)
		return
	}
	Succeed(c, u.Listening)
}

// 新增任务
func AddJob(c *gin.Context) {
	job := user.Job{}
	err := c.Bind(&job)
	if err != nil {
		Failed(c, 1, err.Error(), "received", job)
		return
	}
	u := GetUser(c)
	u.Jobs = append(u.Jobs, job)
	err = u.Update()
	if err != nil {
		Failed(c, 2, err.Error(), "received", u.Jobs)
		return
	}
	Succeed(c, u.Jobs)
}

// 移除任务
func RemoveJobs(c *gin.Context) {
	ids := c.QueryArray("jobs")
	u := GetUser(c)
	err := u.RemoveJobs(ids)
	if err != nil {
		Failed(c, 1, err.Error(), "received", u.Jobs)
		return
	}
	Succeed(c, u.Jobs)
}

// 提交博文
func Submit(c *gin.Context) {
	// 检验数据合法
	if v, ok := c.GetPostForm("repost"); !ok {
		Failed(c, 1, "Form key \"repost\" not found. A \"null\" value must be pass in if there is no repost.", "received", v)
		return
	}
	// 初始化 post
	post := data.Post{Submitter: GetUser(c)}
	// 绑定其他数据
	err := c.Bind(&post)
	if err != nil {
		Failed(c, 2, err.Error(), "received", post)
		return
	}
	// 不允许提交已储存的博文刷积分
	if data.HasPost(post.Platform, post.Mid) {
		Failed(c, 3, "该博文已被提交过", "received", post)
		return
	}
	// 检查该用户是否已提交过
	m := data.GetMonitor(post.Type())
	if m.IsSubmitted(post.Submitter.Uid) {
		Failed(c, 4, "您已提交过", "received", post)
		return
	}
	// 分析去了
	go m.Parse(&post)
	Succeed(c, "提交成功")
}
