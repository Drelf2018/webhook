package api

import (
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/user"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

// 当前版本号
func Version(c *gin.Context) {
	Succeed(c, "v0.4.0")
}

// 查看资源目录
func List(c *gin.Context) {
	Succeed(c, config.Resource)
}

// 解析图片网址并返回文件
// 获取参数 https://blog.csdn.net/weixin_52690231/article/details/124109518
// 返回文件 https://blog.csdn.net/kilmerfun/article/details/123943070
// 重定向至 https://www.ngui.cc/el/3757797.html?action=onClick
func Fetch(c *gin.Context) {
	c.Request.URL.Path = "/" + config.Path.Public + data.Save(c.Param("url")[1:])
	config.Engine.HandleContext(c)
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
	Succeed(c, "auth", auth, "token", token)
}

// 新建用户
func Register(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		auth = c.Query("Authorization")
	}
	u := user.Get(auth)
	if u.Token == "" {
		Failed(c, 1, "请先获取验证码")
		return
	}
	matched, err := u.MatchReplies()
	if err != nil {
		Failed(c, 2, err.Error())
		return
	}
	if !matched {
		Failed(c, 3, "验证失败")
		return
	}
	uid := u.Uid
	user.Done(uid)
	if u.Scan(uid) == nil {
		// 已注册用户
		Succeed(c, u.Token)
	} else {
		// 新建用户
		Succeed(c, u.Make(uid).Token)
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
	Succeed(c, p.Comments)
}
