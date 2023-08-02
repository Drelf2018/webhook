package webhook

import (
	"net/http"
	"time"

	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

// 返回数据
func Succeed(c *gin.Context, data ...any) {
	obj := gin.H{"code": 0}
	if len(data) == 1 {
		obj["data"] = data[0]
	} else if len(data) > 1 {
		temp := gin.H{}
		for i := 0; i < len(data); i += 2 {
			temp[data[i].(string)] = data[i+1]
		}
		obj["data"] = temp
	}
	c.JSON(200, obj)
}

func Failed(c *gin.Context, code int, message string, data ...any) {
	obj := gin.H{"code": code, "message": message}
	for i := 0; i < len(data); i += 2 {
		obj[data[i].(string)] = data[i+1]
	}
	c.AbortWithStatusJSON(200, obj)
}

// 读取用户
func GetUser(c *gin.Context) *user.User {
	return c.MustGet("user").(*user.User)
}

// 验证授权
func Authorize(c *gin.Context) {
	token := c.GetHeader("Authorization")
	user := new(user.User).Query(token)
	if user == nil {
		Failed(c, 1, "登录失败", "received", token)
		return
	}
	Timer.Update(user.Uid)
	c.Set("user", user)
}

// 获取 beginTs 时间之后的所有博文
func GetPosts(c *gin.Context) {
	// 10 秒的冗余还是太短了啊 没事的 10 秒也很厉害了
	TimeNow := time.Now().Unix()
	beginTs := c.Query("begin")
	endTs := c.Query("end")

	begin := utils.Ternary(beginTs != "", beginTs, utils.Time{Stamp: TimeNow - 30}.ToString())
	end := utils.Ternary(endTs != "", endTs, utils.Time{Stamp: TimeNow}.ToString())

	posts := new([]data.Post)
	data.GetPosts(begin, end, posts)

	Succeed(c, "posts", posts, "online", Timer.Update())
}

// 提交博文
func Submit(c *gin.Context) {
	// 初始化 post
	post := data.Post{Submitter: GetUser(c)}
	post.Face.Make(c.PostForm("face"))
	post.Pendant.Make(c.PostForm("pendant"))
	post.Attachments.Make(c.PostFormArray("attachments")...)

	// 绑定其他数据
	err := c.Bind(&post)
	if err != nil {
		Failed(c, 1, err.Error(), "received", post)
		return
	}

	if data.HasPost(post.Platform, post.Mid) {
		Failed(c, 2, "该博文已被提交过", "received", post)
		return
	}

	m := Monitors.Get(post.Platform + post.Mid)

	if m.In(post.Submitter.Uid) {
		Failed(c, 3, "您已提交过", "received", post)
		return
	}

	go m.Parse(&post)
	Succeed(c, "提交成功")
}

// 获取所有分支
func GetBranches(c *gin.Context) {
	platform := c.Param("platform")
	mid := c.Param("mid")
	r := new([]data.Post)
	data.GetBranches(platform, mid, r)
	Succeed(c, r)
}

// 获取所有评论
func GetComments(c *gin.Context) {
	platform := c.Param("platform")
	mid := c.Param("mid")
	p := data.GetPost(platform, mid)
	if p == nil {
		Failed(c, 1, "未找到评论")
		return
	}
	cs := data.Comments{Root: p}
	cs.Query()
	Succeed(c, cs.Root)
}

// 解决跨域问题
//
// 参考: https://blog.csdn.net/u011866450/article/details/126958238
func Cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")

	// 禁止所有 OPTIONS 方法 原因见博文
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func Run(cfg *Config) {
	// 自动填充
	cfg.AutoFill()

	// 设置运行模式
	gin.SetMode(utils.Ternary(cfg.Debug, gin.DebugMode, gin.ReleaseMode))

	r := gin.Default()

	// 跨域设置
	r.Use(Cors)

	// 资源文件相关
	data.Reset(cfg.Resource, cfg.File)
	r.Static(cfg.Resource, cfg.Resource)
	r.StaticFile("favicon.ico", cfg.Resource+"/favicon.ico")

	// 解析图片网址并返回文件
	// 获取参数 https://blog.csdn.net/weixin_52690231/article/details/124109518
	// 返回文件 https://blog.csdn.net/kilmerfun/article/details/123943070
	r.GET("fetch/*u", func(c *gin.Context) { c.File(cfg.Resource + new(data.Attachment).Make(c.Param("u")[1:]).Path) })

	// 获取当前在线状态
	r.GET("/online", func(c *gin.Context) { Succeed(c, Timer.Update()) })

	// 新建用户
	r.GET("/user/new", func(c *gin.Context) { Succeed(c, new(user.User).Make(c.Query("uid")).Token) })

	r.GET("/posts", GetPosts)

	r.GET("/branches/:platform/:mid", GetBranches)

	r.GET("/comments/:platform/:mid", GetComments)

	// 以下操作需要通过登录验证
	r.Use(Authorize)

	// 更新自身在线状态
	r.GET("/ping", func(c *gin.Context) { Timer.Update(GetUser(c).Uid) })

	// 获取自身信息
	r.GET("/user/me", func(c *gin.Context) { Succeed(c, GetUser(c)) })

	// 提交博文
	r.POST("/submit", Submit)

	// 注册 获取token 修改用户信息 提交配置信息

	// 运行 gin 服务器 默认 0.0.0.0:9000
	r.Run(cfg.Addr())
}
