package api

import (
	"runtime"

	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type Visitor int

// 当前版本号
func (Visitor) GetVersion(c *gin.Context) {
	Succeed(c, config.VERSION)
}

func Ternary[K comparable, T any](expr K, value K, a, b T) T {
	if expr == value {
		return a
	}
	return b
}

// 查看资源目录
func (Visitor) GetList(c *gin.Context) {
	cmd := Ternary(runtime.GOOS, "windows", "dir /s /b", "du -ah")
	s, err := Shell(cmd)
	if err != nil {
		Abort(c, err)
		return
	}
	Succeed(c, s)
}

// 获取当前在线状态
func (Visitor) GetOnline(c *gin.Context) {
	Succeed(c, utils.Timer())
}

var tokens = make(map[string][2]string)

// 获取随机 Token
func generateRandomToken(uid string) (auth, token string) {
	token = utils.RandomNumberMixString(6, 3)
	auth = uuid.NewV4().String()
	tokens[uid] = [2]string{auth, token}
	return
}

func getDataByAuth(auth string) (uid string, token string, ok bool) {
	for uid, data := range tokens {
		if data[0] == auth {
			return uid, data[1], true
		}
	}
	return "", "", false
}

// 获取注册所需 Token
func (Visitor) GetToken(c *gin.Context) {
	uid := c.Query("uid")
	if !u20.IsDigit(uid) {
		Abort(c, "请正确填写纯数字的 uid 参数")
		return
	}
	auth, token := generateRandomToken(uid)
	Succeed(c, "auth", auth, "token", token, "oid", config.Global.Oid)
}

// 新建用户
func (Visitor) GetRegister(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		auth = c.Query("Authorization")
	}
	uid, token, ok := getDataByAuth(auth)
	if !ok {
		Abort(c, "请先获取验证码")
		return
	}
	matched, err := utils.SearchToken(uid, token)
	if err != nil {
		Abort(c, err)
		return
	}
	if !matched {
		Abort(c, "验证失败")
		return
	}
	delete(tokens, uid)
	if u := model.QueryAuth(uid); u != nil {
		// 已注册用户
		Succeed(c, u.Auth)
	} else {
		// 新建用户
		Succeed(c, model.NewUser(uid).Auth)
	}
}

// 获取 begin 与 end 时间范围内的所有博文
func (Visitor) GetPosts(c *gin.Context) {
	if _, ok := c.GetQuery("test"); ok {
		Succeed(c, "posts", []model.Post{*model.TestPost}, "online", utils.Timer())
		return
	}
	begin, end := c.Query("begin"), c.Query("end")
	TimeNow := utils.NewTime(nil)
	if end == "" {
		end = TimeNow.ToString()
	}
	if begin == "" {
		// 10 秒的冗余还是太短了啊 没事的 10 秒也很厉害了
		begin = TimeNow.Delay(-30).ToString()
	}
	Succeed(c, "posts", model.GetPosts(begin, end), "online", utils.Timer())
}

// 获取所有分支
// func GetBranches(c *gin.Context) {
// 	platform, mid := c.Param("platform"), c.Param("mid")
// 	posts := make([]data.Post, 0)
// 	data.GetBranches(platform, mid, &posts)
// 	Succeed(c, posts)
// }

// 获取所有评论
// func GetComments(c *gin.Context) {
// 	platform, mid := c.Param("platform"), c.Param("mid")
// 	p := data.GetPost(platform, mid)
// 	if p == nil {
// 		Failed(c, 1, "未找到评论")
// 		return
// 	}
// 	Succeed(c, p.Comments)
// }

// func (Visitor) GetParse(c *gin.Context) {
// 	Succeed(c, "error", model.TestPost.Parse())
// }
