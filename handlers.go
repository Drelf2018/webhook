package webhook

import (
	"github.com/Drelf2018/webhook/data"
	"github.com/Drelf2018/webhook/user"
	"github.com/Drelf2018/webhook/utils"
	utils20 "github.com/Drelf2020/utils"
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

	m := data.GetMonitor(post.Platform + post.Mid)

	if m.Has(post.Submitter.Uid) {
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

// 生成随机验证码
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
