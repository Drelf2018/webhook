package webhook

import (
	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type LifeCycle interface {
	// 初始化
	OnCreate(*Config)
	// 跨域设置
	OnCors(*Config)
	// 静态资源绑定
	OnStatic(*Config)
	// 无鉴权接口
	BeforeAuthorize(*Config)
	// 鉴权
	OnAuthorize(*Config)
	// 有鉴权接口
	AfterAuthorize(*Config)
	// 鉴定管理员权限
	OnAdmin(*Config)
	// 管理员接口
	AfterAdmin(*Config)
}

type Cycle struct{}

func (c Cycle) OnCreate(r *Config) {
	// 多次尝试克隆主页到本地
	go utils.RetryError(10, 0, r.IndexUpdate)
	// 资源数据库初始化
	data.Connect(r.ToPublic(), r.Public.Path, r.ToPostsDB())
	// 用户数据库初始化
	user.Connect(r.ToUsersDB())
	// 指定查询网址
	user.MakeUrl("643451139714449427")
}

func (c Cycle) OnCors(r *Config) {
	r.Use(Cors)
}

func (c Cycle) OnStatic(r *Config) {
	// 主页绑定
	r.Use(static.ServeRoot("/", r.ToIndex()))
	// 静态资源绑定
	r.Static(r.Public.Path, r.ToPublic())
}

func (c Cycle) BeforeAuthorize(r *Config) {
	// 主动更新主页
	r.GET("/update", func(c *gin.Context) {
		err := r.IndexUpdate()
		if err != nil {
			Failed(c, 1, err.Error(), "folder", r.ToIndex())
			return
		}
		Succeed(c)
	})

	r.GET("/list", func(c *gin.Context) { Succeed(c, r.List()) })

	// 解析图片网址并返回文件
	r.GET("/fetch/*url", FetchFile, func(c *gin.Context) { r.HandleContext(c) })

	// 获取当前在线状态
	r.GET("/online", func(c *gin.Context) { Succeed(c, utils.Timer()) })

	// 获取注册所需 Token
	r.GET("/token", GetToken)

	// 新建用户
	r.GET("/register", Register)

	r.GET("/posts", GetPosts)

	r.GET("/branches/:platform/:mid", GetBranches)

	r.GET("/comments/:platform/:mid", GetComments)
}

func (c Cycle) OnAuthorize(r *Config) {
	r.Use(Authorize)
}

func (c Cycle) AfterAuthorize(r *Config) {
	// 更新自身在线状态
	r.GET("/ping", func(c *gin.Context) { utils.Timer(GetUser(c).Uid) })

	// 获取自身信息
	r.GET("/user/me", func(c *gin.Context) { Succeed(c, GetUser(c)) })

	// 提交博文
	r.POST("/submit", Submit)

	// 提权
	// r.GET("/whosyourdaddy")
	// 修改用户信息 提交配置信息 待实现
}

func (c Cycle) OnAdmin(r *Config) {
}

func (c Cycle) AfterAdmin(r *Config) {
}
