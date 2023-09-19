package service

import (
	"fmt"
	"net/http"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/api"
	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type Cycle int

func (c Cycle) OnCreate(r *configs.Config) {
	// 资源数据库初始化
	data.SetSqlite(data.Public().Path(r.Path.Posts))
	// 用户数据库初始化
	user.SetOid("643451139714449427")
	user.SetSqlite(r.Resource.Path(r.Path.Users))
	if gin.Mode() == gin.DebugMode {
		user.CreateTestUser()
	}
	// 多次尝试克隆主页到本地
	go asyncio.RetryError(10, 0, r.UpdateIndex)
}

// 解决跨域问题
//
// 参考: https://blog.csdn.net/u011866450/article/details/126958238
func (c Cycle) OnCors(r *configs.Config) {
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		// 禁止所有 OPTIONS 方法 原因见博文
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
		}
	})
}

func (c Cycle) OnStatic(r *configs.Config) {
	index := r.Resource.Path(r.Path.Views)
	// 主页绑定
	r.Use(static.ServeRoot("/", index))
	// 子页面
	r.Use(static.ServeRoot("/user", index))
	// 静态资源绑定
	r.Static(r.Path.Public, data.Public().Path())
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

	// 提交博文
	r.POST("/submit", api.Submit)

	// 修改用户信息 提交配置信息 待实现
}

func (c Cycle) OnAdmin(r *configs.Config) {
	r.Use(api.IsAdministrator)
}

func (c Cycle) Administrator(r *configs.Config) {
	// 在资源文件架执行命令
	// r.GET("/exec/*cmd", api.Cmd)
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
