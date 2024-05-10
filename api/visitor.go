package api

import (
	"runtime"
	"slices"

	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/database/dao"
	"github.com/Drelf2018/webhook/database/model"
	"github.com/Drelf2018/webhook/interfaces"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"

	group "github.com/Drelf2018/gin-group"
)

var Visitor = group.Group{
	Handlers: group.Chain{
		GetVersion,
		GetList,
		GetOnline,
		GetToken,
		GetRegister,
	},
}

// 当前版本号
func GetVersion(ctx *gin.Context) (any, group.Error) {
	return config.VERSION, nil
}

var CmdList string

func init() {
	switch runtime.GOOS {
	case "windows":
		CmdList = "dir /s /b"
	case "linux":
		CmdList = "du -ah"
	}
}

// 查看资源目录
func GetList(ctx *gin.Context) (any, group.Error) {
	return WrapError(utils.RunShell(CmdList, ""))
}

// 获取当前在线状态
func GetOnline(ctx *gin.Context) (any, group.Error) {
	return utils.OnlineList, nil
}

// 获取注册所需 Token
func GetToken(ctx *gin.Context) (any, group.Error) {
	return WrapError(interfaces.Token(ctx))
}

// 新建用户
func GetRegister(ctx *gin.Context) (any, group.Error) {
	uid, err := interfaces.Register(ctx)
	if err != nil {
		return nil, group.AutoError(err)
	}

	user := dao.QueryUserByUID(uid)

	// 已注册
	if user != nil {
		return user.Auth, nil
	}

	// 新建用户
	var permission model.Permission
	switch {
	case uid == config.Global.Permission.Owner:
		permission = model.Owner
	case slices.Contains(config.Global.Permission.Administrators, uid):
		permission = model.Administrator
	case slices.Contains(config.Global.Permission.Trustors, uid):
		permission = model.Trustor
	}
	return dao.NewUser(uid, permission).Auth, nil
}

// 获取 begin 与 end 时间范围内的所有博文
// func GetPosts(ctx *gin.Context) (any, group.Error) {
// 	if _, ok := ctx.GetQuery("test"); ok {
// 		Succeed(c, "posts", []model.Post{*model.TestPost}, "online", utils.Timer())
// 		return
// 	}
// 	begin, end := c.Query("begin"), c.Query("end")
// 	TimeNow := utils.NewTime(nil)
// 	if end == "" {
// 		end = TimeNow.ToString()
// 	}
// 	if begin == "" {
// 		// 10 秒的冗余还是太短了啊 没事的 10 秒也很厉害了
// 		begin = TimeNow.Delay(-30).ToString()
// 	}
// 	Succeed(c, "posts", dao.GetPosts(begin, end), "online", utils.Timer())
// }

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
