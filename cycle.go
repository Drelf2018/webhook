package webhook

import (
	"net/http"
	"strings"

	"github.com/Drelf2018/asyncio"
	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"slices"
)

type LifeCycle interface {
	// 初始化
	OnCreate(*Config)
	// 跨域设置
	OnCors(*Config)
	// 静态资源绑定
	OnStatic(*Config)
	// 访客接口
	Visitor(*Config)
	// 鉴定提交者权限
	OnAuthorize(*Config)
	// 提交者接口
	Submitter(*Config)
	// 鉴定管理员权限
	OnAdmin(*Config)
	// 管理员接口
	Administrator(*Config)
}

type Cycle int

func (c Cycle) OnCreate(r *Config) {
	// 资源数据库初始化
	data.Connect(r.ToPublic(), r.Public.Path, r.ToPostsDB())
	// 用户数据库初始化
	user.Connect("643451139714449427", r.ToUsersDB())
	// 多次尝试克隆主页到本地
	go asyncio.RetryError(10, 0, r.IndexUpdate)
}

// 解决跨域问题
//
// 参考: https://blog.csdn.net/u011866450/article/details/126958238
func (c Cycle) OnCors(r *Config) {
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

func (c Cycle) OnStatic(r *Config) {
	// 主页绑定
	r.Use(static.ServeRoot("/", r.ToIndex()))
	// 静态资源绑定
	r.Static(r.Public.Path, r.ToPublic())
}

func (c Cycle) Visitor(r *Config) {
	// 版本
	r.GET("/version", func(c *gin.Context) { Succeed(c, "v0.3.0") })
	// 查看资源目录
	r.GET("/list", func(c *gin.Context) { Succeed(c, r.List()) })

	// 解析图片网址并返回文件
	// 获取参数 https://blog.csdn.net/weixin_52690231/article/details/124109518
	// 返回文件 https://blog.csdn.net/kilmerfun/article/details/123943070
	// 重定向至 https://www.ngui.cc/el/3757797.html?action=onClick
	r.GET("/fetch/*url", func(c *gin.Context) {
		c.Request.URL.Path = data.NewA(c.Param("url")[1:]).ToURL()
		r.HandleContext(c)
	})

	// 获取当前在线状态
	r.GET("/online", func(c *gin.Context) { Succeed(c, utils.Timer()) })

	// 获取注册所需 Token
	r.GET("/token", GetToken)

	// 新建用户
	r.GET("/register", Register)

	// 查询博文
	r.GET("/posts", GetPosts)

	// 查询某博文不同版本
	r.GET("/branches/:platform/:mid", GetBranches)

	// 查询某博文的评论
	r.GET("/comments/:platform/:mid", GetComments)
}

func (c Cycle) OnAuthorize(r *Config) {
	r.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("Authorization")
		}
		if token == "" {
			Failed(c, 1, "你是不是调错接口了啊")
			return
		}
		user := new(user.User).Query(token)
		if user == nil {
			Failed(c, 2, "鉴权失败", "received", token)
			return
		}
		utils.Timer(user.Uid)
		c.Set("user", user)
	})
}

func (c Cycle) Submitter(r *Config) {
	// 更新自身在线状态
	r.GET("/ping", func(c *gin.Context) { utils.Timer(GetUser(c).Uid) })

	// 获取自身信息
	r.GET("/me", func(c *gin.Context) { Succeed(c, GetUser(c)) })

	// 主动更新主页
	r.GET("/update", func(c *gin.Context) {
		err := r.IndexUpdate()
		if err != nil {
			Failed(c, 1, err.Error())
			return
		}
		Succeed(c)
	})

	// 提交博文
	r.POST("/submit", Submit)

	// 修改用户信息 提交配置信息 待实现
}

func (c Cycle) OnAdmin(r *Config) {
	r.Use(func(c *gin.Context) {
		user := GetUser(c)
		if !slices.Contains(r.Administrators, user.Uid) {
			Failed(c, 1, "您没有管理员权限")
		}
	})
}

func (c Cycle) Administrator(r *Config) {
	// 在资源文件架执行命令
	r.GET("/exec/*cmd", func(c *gin.Context) {
		cmds := strings.Split(c.Param("cmd")[1:], "/")
		err := r.Exec(cmds)
		if err != nil {
			Failed(c, 1, err.Error())
			return
		}
		Succeed(c)
	})
}
