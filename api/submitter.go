package api

import (
	"errors"
	"strings"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook/database/dao"
	"github.com/Drelf2018/webhook/database/model"
	"github.com/Drelf2018/webhook/utils"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

var log *logrus.Logger

var (
	ErrNoAuthCode = errors.New("webhook/api: no authentication code")
	ErrAuth       = errors.New("webhook/api: authentication failure")
)

var Submitter = group.Group{
	Path: "submitter",
	// Middlewares: gin.HandlersChain{SetUser},
	Handlers: group.Chain{
		GetMe,
		GetPing,
		GetRemove_job,
		GetModify_follow,
		PostSubmit,
		PostAdd_job,
		PostTest_job,
	},
}

func SetUser(ctx *gin.Context) {
	// 从 headers 或者 query 获取身份码
	token := ctx.GetHeader("Authorization")
	if token == "" {
		token = ctx.Query("auth")
	}
	if token == "" {
		token = ctx.Query("Auth")
	}
	if token == "" {
		token = ctx.Query("authorization")
	}
	if token == "" {
		token = ctx.Query("Authorization")
	}
	if token == "" {
		group.Abort(ctx, nil, group.AutoError(ErrNoAuthCode))
		return
	}

	user := dao.QueryUserByAuth(token)
	if user == nil {
		group.Abort(ctx, nil, group.AutoError(ErrAuth))
		return
	}

	// 修改在线时间戳
	utils.OnlineList.Update(user.Uid)
	group.SetUser(ctx, *user)

	// 打印日志 不记录 ping 请求
	if !strings.Contains(ctx.Request.URL.Path, "/ping") {
		log.Infof("%v %v \"%v\"", user, ctx.Request.Method, ctx.Request.URL)
	}
}

// 更新在线时间
func GetPing(ctx *gin.Context) (any, group.Error) {
	return "pong", nil
}

// 获取自身信息
func GetMe(ctx *gin.Context) (any, group.Error) {
	return group.GetUser[model.User](ctx), nil
}

// 更新监听列表
func GetModify_follow(ctx *gin.Context) (any, group.Error) {
	u := group.GetUser[model.User](ctx)
	u.Follow = ctx.QueryArray("follow")
	err := dao.UpdateUser(&u)
	if err != nil {
		return nil, group.AutoError(err)
	}
	return u.Follow, nil
}

// 新增任务
func PostAdd_job(ctx *gin.Context) (any, group.Error) {
	job := model.Job{}
	err := ctx.Bind(&job)
	if err != nil {
		return nil, group.AutoError(err)
	}
	u := group.GetUser[model.User](ctx)
	u.Jobs = append(u.Jobs, job)
	err = dao.UpdateUser(&u)
	if err != nil {
		return nil, group.AutoError(err)
	}
	return u.Jobs, nil
}

// 移除任务
func GetRemove_job(ctx *gin.Context) (any, group.Error) {
	u := group.GetUser[model.User](ctx)
	err := dao.RemoveJobs(&u, ctx.QueryArray("jobs"))
	if err != nil {
		return nil, group.AutoError(err)
	}
	return u.Jobs, nil
}

// 测试单个任务
func PostTest_job(ctx *gin.Context) (any, group.Error) {
	var job model.Job
	err := ctx.Bind(&job)
	if err != nil {
		return nil, group.AutoError(err)
	}
	return job.Send(model.TestPost), nil
}

// 测试任务
// func (Submitter) GetTests(c *gin.Context) {
// 	jobs := dao.GetJobsByID(User(c).Uid, c.QueryArray("jobs"))
// 	model.TestPost.ScanAndSend()
// 	Succeed(c, model.TestPost.Send(jobs))
// }

var ErrNoRepost = group.NewError(1, `Form key "repost" not found. A "null" value must be pass in if there is no repost.`)

// 提交博文
func PostSubmit(ctx *gin.Context) (any, group.Error) {
	// 检验数据合法
	v, ok := ctx.GetPostForm("repost")
	if !ok {
		return v, ErrNoRepost
	}
	// 初始化 post
	u := group.GetUser[model.User](ctx)
	post := model.Post{Submitter: &u}
	// 绑定其他数据
	err := ctx.Bind(&post)
	if err != nil {
		return nil, group.AutoError(err)
	}
	// 不允许提交已储存的博文刷积分
	if dao.ExistsPost(&post) {
		return nil, group.NewError(144, "该博文已被提交过")
	}
	dao.SavePost(&post)
	dao.Webhook(&post)
	return "提交成功", nil
}
