package api

import (
	"net/http"
	"os"
	"runtime"

	"github.com/Drelf2018/webhook/configs"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

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

// 当前版本号
func Version(c *gin.Context) {
	Succeed(c, "v0.9.1")
}

// 查看资源目录
func List(c *gin.Context) {
	cmd := u20.Ternary(runtime.GOOS == "windows", "dir /s /b", "du -ah")
	s, err := Shell(cmd)
	Final(c, 1, err, nil, s)
}

// 解析图片网址并返回文件
//
// 获取参数 https://blog.csdn.net/weixin_52690231/article/details/124109518
//
// 返回文件 https://blog.csdn.net/kilmerfun/article/details/123943070
//
// 重定向至 https://www.ngui.cc/el/3757797.html?action=onClick
func Fetch(c *gin.Context) {
	c.Request.URL.Path = data.Save(c.Param("url")[1:])
	configs.Get().Engine.HandleContext(c)
}

// 获取当前在线状态
func Online(c *gin.Context) {
	Succeed(c, utils.Timer())
}

// 获取注册所需 Token
func GetToken(c *gin.Context) {
	uid := c.Query("uid")
	if !u20.IsDigit(uid) {
		Failed(c, 1, "请正确填写纯数字的 uid 参数", "received", uid)
		return
	}
	auth, token := user.GetRandomToken(uid)
	Succeed(c, "auth", auth, "token", token, "oid", configs.Get().Oid)
}

// 新建用户
func Register(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		auth = c.Query("Authorization")
	}
	u, ok := user.Tokens[auth]
	if !ok {
		Failed(c, 1, "请先获取验证码")
		return
	}
	matched, err := u.MatchReplies()
	if Error(c, 2, err) {
		return
	}
	if !matched {
		Failed(c, 3, "验证失败")
		return
	}
	user.Done(u.Uid)
	if user.Users.First(&u, "uid = ?", u.Uid) {
		// 已注册用户
		Succeed(c, u.Token)
	} else {
		// 新建用户
		Succeed(c, user.Make(u.Uid).Token)
	}
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

// 获取所有分支
func GetBranches(c *gin.Context) {
	platform, mid := c.Param("platform"), c.Param("mid")
	posts := make([]data.Post, 0)
	data.GetBranches(platform, mid, &posts)
	Succeed(c, posts)
}

// 获取所有评论
func GetComments(c *gin.Context) {
	platform, mid := c.Param("platform"), c.Param("mid")
	p := data.GetPost(platform, mid)
	if p == nil {
		Failed(c, 1, "未找到评论")
		return
	}
	Succeed(c, p.Comments)
}

// 读取日志
func ReadLog(c *gin.Context) {
	b, err := os.ReadFile(configs.Get().Path.Full.Log)
	Final(c, 1, err, nil, CutString(b))
}
