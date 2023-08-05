package webhook

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/network"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	utils20 "github.com/Drelf2020/utils"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// 返回成功数据
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

// 返回错误信息
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
}

// 获取 begin 与 end 时间范围内的所有博文
func GetPosts(c *gin.Context) {
	begin, end := c.Query("begin"), c.Query("end")
	TimeNow := utils.NewTime(nil)
	if end == "" {
		end = TimeNow.ToString()
	}
	if begin == "" {
		// 10 秒的冗余还是太短了啊 没事的 10 秒也很厉害了
		begin = TimeNow.Delay(-30).ToString()
	}
	posts := make([]data.Post, 0)
	data.GetPosts(begin, end, &posts)
	Succeed(c, "posts", posts, "online", utils.Timer())
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
	posts := make([]data.Post, 0)
	data.GetBranches(platform, mid, &posts)
	Succeed(c, posts)
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

func GetToken(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" || !utils20.IsNumber(uid) {
		Failed(c, 1, "请正确填写纯数字的 uid 参数", "received", uid)
		return
	}
	auth, token := user.GetRandomToken(uid)
	Succeed(c, "auth", auth, "token", token)
}

// 注册
func Register(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	u := user.Get(auth)
	if u.Token == "" {
		Failed(c, 1, "请先获取验证码")
		return
	}
	matched, err := network.MatchReplies(u.Uid, u.Token)
	if err != nil {
		Failed(c, 2, err.Error())
		return
	}
	if !matched {
		Failed(c, 3, "验证失败")
		return
	}
	user.Done(u.Uid)
	Succeed(c, u.Make(u.Uid).Token)
}

// 获取参数 https://blog.csdn.net/weixin_52690231/article/details/124109518
// 返回文件 https://blog.csdn.net/kilmerfun/article/details/123943070
// 重定向至 https://www.ngui.cc/el/3757797.html?action=onClick
func FetchFile(c *gin.Context) {
	c.Request.URL.Path = data.NewA(c.Param("url")[1:]).ToURL()
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

// 主页更新
func IndexUpdate(g network.Github) bool {
	// 先判断存不存在
	if utils.FileNotExists(g.Repository) {
		// 再决定要不要克隆
		exec.Command("git", "clone", "-b", g.Branche, g.ToGIT()).Run()
		return false
	}
	// 存在则判断版本
	if g.GetLastCommit() != nil {
		return false
	}
	// 读取目录下版本信息
	if g.NoVersion() {
		return g.Write()
	}
	// 版本不正确直接删库重来
	if !g.Check() {
		os.RemoveAll(g.Repository)
		return false
	}
	return true
}

// 鉴权前
func BeforeAuthorize(r *Config) {
	// 解析图片网址并返回文件
	r.GET("/fetch/*url", FetchFile, func(c *gin.Context) { r.HandleContext(c) })

	// 获取当前在线状态
	r.GET("/online", func(c *gin.Context) { Succeed(c, utils.Timer()) })

	// 指定查询网址
	network.MakeUrl("643451139714449427")

	// 获取注册所需 Token
	r.GET("/token", GetToken)

	// 新建用户
	r.GET("/register", Register)

	r.GET("/posts", GetPosts)

	r.GET("/branches/:platform/:mid", GetBranches)

	r.GET("/comments/:platform/:mid", GetComments)
}

// 鉴权后
func AfterAuthorize(r *Config) {
	// 主动更新主页
	r.GET("/update", func(c *gin.Context) { Succeed(c, IndexUpdate(r.Github)) })

	// 更新自身在线状态
	r.GET("/ping", func(c *gin.Context) { utils.Timer(GetUser(c).Uid) })

	// 获取自身信息
	r.GET("/user/me", func(c *gin.Context) { Succeed(c, GetUser(c)) })

	// 提交博文
	r.POST("/submit", Submit)

	// 修改用户信息 提交配置信息 待实现
}

func Run(r *Config) {
	// 自动填充配置
	r.AutoFill()
	// 载入 handlers
	if r.DIY != nil {
		r.DIY(r)
	} else {
		// 跨域设置
		r.Use(Cors)
		// 多次尝试克隆主页到本地
		go utils.Retry[network.Github](10, 10, IndexUpdate, r.Github)
		// 主页绑定
		r.Use(static.ServeRoot("/", "./"+r.Github.Repository))
		// 资源数据库初始化
		data.Connect(r.Resource, r.File)
		// 静态资源绑定
		r.Static(r.Resource, r.Resource)
		// 具体接口实现
		r.BeforeAuthorize(r)
		r.Use(Authorize)
		r.AfterAuthorize(r)
	}
	// 运行 gin 服务器 默认 0.0.0.0:9000
	r.Engine.Run(r.Url + ":" + r.Port)
}
