package api

import (
	"strings"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

var log = u20.Logger()

// 设置值类型对象默认值
func SetZero[T comparable](a *T, b ...T) {
	var zero T
	if *a == zero {
		for _, c := range b {
			if c == zero {
				continue
			}
			*a = c
			break
		}
	}
}

func IsSubmitter(c *gin.Context) {
	// 从 headers 或者 query 获取身份码
	token := c.GetHeader("Authorization")
	token1, ok1 := c.GetQuery("Authorization")
	token2, ok2 := c.GetQuery("authorization")
	SetZero(&token, token1, token2)

	if token == "" {
		Failed(c, 1, "需要身份鉴权")
		return
	}

	u := user.Query(token)
	if u == nil {
		Failed(c, 2, "鉴权失败", "received", token)
		return
	}

	// 修改在线时间戳
	utils.Timer(u.Uid)
	c.Set("user", u)

	// 不记录 ping 请求
	if strings.Contains(c.Request.URL.Path, "/ping") {
		return
	}

	// 清除 query 中的身份码
	if ok1 || ok2 {
		query := make([]string, 0)
		for k, v := range c.Request.URL.Query() {
			if k == "Authorization" || k == "authorization" {
				continue
			}
			query = append(query, k+"="+strings.Join(v, "&"+k+"="))
		}
		c.Request.URL.RawQuery = strings.Join(query, "&")
	}
	// 打印日志
	log.Infof("%v %v \"%v\"", u, c.Request.Method, c.Request.URL)
}

// 更新在线时间
func Ping(c *gin.Context) {
	utils.Timer(GetUser(c).Uid)
}

// 获取自身信息
func Me(c *gin.Context) {
	Succeed(c, c.MustGet("user"))
}

// 主动更新主页
func Update(c *gin.Context) {
	cfg := configs.Get()
	Final(c, 1, cfg.UpdateIndex(), nil, cfg.Github.Commit.Sha)
}

// 更新监听列表
func ModifyListening(c *gin.Context) {
	u := GetUser(c)
	u.Listening = c.QueryArray("listen")
	Final(c, 1, u.Update(), []any{"received", u.Listening}, u.Listening)
}

// 新增任务
func AddJob(c *gin.Context) {
	job := user.Job{}
	if Error(c, 1, c.Bind(&job), "received", job) {
		return
	}
	u := GetUser(c)
	u.Jobs = append(u.Jobs, job)
	Final(c, 2, u.Update(), []any{"received", u.Jobs}, u.Jobs)
}

// 移除任务
func RemoveJobs(c *gin.Context) {
	u := GetUser(c)
	Final(c, 1, u.RemoveJobs(c.QueryArray("jobs")), []any{"received", u.Jobs}, u.Jobs)
}

// 测试单个任务
func TestJob(c *gin.Context) {
	job := user.Job{}
	if Error(c, 1, c.Bind(&job), "received", job) {
		return
	}
	Succeed(c, data.TestPost.Send([]user.Job{job})[0])
}

// 测试任务
func TestJobs(c *gin.Context) {
	Succeed(c, data.TestPost.Send(user.GetJobsByID(GetUser(c).Uid, c.QueryArray("jobs")...)))
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
	if Error(c, 2, c.Bind(&post), "received", post) {
		return
	}
	// 依赖注入
	initial.Default(&post)
	// 不允许提交已储存的博文刷积分
	if data.HasPost(post.Platform, post.Mid) {
		Failed(c, 3, "该博文已被提交过", "received", post)
		return
	}
	// 检查该用户是否已提交过
	Final(c, 4, post.Parse(), []any{"received", post}, "提交成功")
}
