package service

import (
	"fmt"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2020/utils"
	"github.com/gin-contrib/static"
)

type LifeCycle interface {
	// 初始化
	OnCreate(*configs.Config)
	// 跨域设置
	OnCors(*configs.Config)
	// 静态资源绑定
	OnStatic(*configs.Config)
	// 访客接口
	Visitor(*configs.Config)
	// 鉴定提交者权限
	OnAuthorize(*configs.Config)
	// 提交者接口
	Submitter(*configs.Config)
	// 鉴定管理员权限
	OnAdmin(*configs.Config)
	// 管理员接口
	Administrator(*configs.Config)
	// 绑定所有接口
	Bind(*configs.Config)
}

type Cycle int

func (c Cycle) OnCreate(r *configs.Config) {
	// 日志初始化
	utils.SetOutputFile(r.Path.Full.Log)
	utils.SetTimestampFormat("2006-01-02 15:04:05")
	// 用户数据库初始化
	user.Init(r)
	// 资源数据库初始化
	data.Init(r)
	// 多次尝试克隆主页到本地
	go asyncio.RetryError(10, 0, r.UpdateIndex)
	// 检查文件是否存在
	go data.CheckFiles()
}

// 解决跨域问题
//
// 参考: https://blog.csdn.net/u011866450/article/details/126958238
func (c Cycle) OnCors(r *configs.Config) {
	r.Use(api.Cors)
}

func (c Cycle) OnStatic(r *configs.Config) {
	index := r.Path.Full.Views
	// 主页绑定
	r.Use(static.ServeRoot("/", index))
	// 子页面
	r.Use(static.ServeRoot("/user", index))
	// 静态资源绑定
	r.Static("/"+r.Path.Public, r.Path.Full.Public)
}

func (c Cycle) Visitor(r *configs.Config) {
	// 版本
	r.GET("/version", api.Version)

	// 查看资源目录
	r.GET("/list", api.List)

	// 解析图片网址并返回文件
	r.GET("/fetch/*url", api.Fetch)

	// 获取当前在线状态
	r.GET("/online", api.Online)

	// 获取注册所需 Token
	r.GET("/token", api.GetToken)

	// 新建用户
	r.GET("/register", api.Register)

	// 查询博文
	r.GET("/posts", api.GetPosts)

	// 查询某博文不同版本
	r.GET("/branches/:platform/:mid", api.GetBranches)

	// 查询某博文的评论
	r.GET("/comments/:platform/:mid", api.GetComments)

	// 读取日志
	r.GET("/log", api.ReadLog)
}

func (c Cycle) OnAuthorize(r *configs.Config) {
	r.Use(api.IsSubmitter)
}

func (c Cycle) Submitter(r *configs.Config) {
	// 更新自身在线状态
	r.GET("/ping", api.Ping)

	// 获取自身信息
	r.GET("/me", api.Me)

	// 主动更新主页
	r.GET("/update", api.Update)

	// 更新监听列表
	r.GET("/modify", api.ModifyListening)

	// 新增任务
	r.POST("/add", api.AddJob)

	// 移除任务
	r.GET("/remove", api.RemoveJobs)

	// 测试单个任务
	r.POST("/test", api.TestJob)

	// 测试任务
	r.GET("/tests", api.TestJobs)

	// 提交博文
	r.POST("/submit", api.Submit)
}

func (c Cycle) OnAdmin(r *configs.Config) {
	r.Use(api.IsAdministrator)
}

func (c Cycle) Administrator(r *configs.Config) {
	// 在资源文件架执行命令
	r.GET("/exec/*cmd", api.Cmd)

	// 获取所有用户信息
	r.GET("/users", api.Users)

	// 修改用户权限
	r.GET("/permission/:uid/:permission", api.UpdatePermission)

	// 结束进程
	r.GET("/close", api.Close)

	// 删除资源
	r.GET("/clear", api.Clear)

	// 删库重启
	r.GET("/reboot", api.Reboot)

	// 检查资源
	r.GET("/files", api.CheckFiles)
}

func (c Cycle) Bind(r *configs.Config) {
	// 初始化
	c.OnCreate(r)
	// 跨域设置
	c.OnCors(r)
	// 静态资源绑定
	c.OnStatic(r)
	// 访客接口
	c.Visitor(r)
	// 鉴定提交者权限
	c.OnAuthorize(r)
	// 提交者接口
	c.Submitter(r)
	// 鉴定管理员权限
	c.OnAdmin(r)
	// 管理员接口
	c.Administrator(r)
	// 运行 gin 服务器 默认 0.0.0.0:9000
	r.Run(fmt.Sprintf(":%v", r.Port))
}
