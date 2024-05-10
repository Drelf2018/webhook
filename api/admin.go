package api

import (
	"errors"
	"os"
	"strconv"

	"github.com/Drelf2018/asyncio"
	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/database/dao"
	"github.com/Drelf2018/webhook/database/model"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

var (
	ErrNoAdmin = errors.New("webhook/api: no administrator permission")
)

var Admin = group.Group{
	Path:        "admin",
	Middlewares: gin.HandlersChain{IsAdmin},
	Handlers: group.Chain{
		GetExec,
		GetGet_users,
		GetSet_permission,
		GetLog,
		GetRestart,
		GetClear_public,
		GetClear_root,
		GetUpdate,
		GetDownload_users,
	},
}

func IsAdmin(ctx *gin.Context) {
	if !group.GetUser[model.User](ctx).IsAdmin() {
		group.Abort(ctx, nil, group.AutoError(ErrNoAdmin))
	}
}

// 主动更新主页
func GetUpdate(ctx *gin.Context) (any, group.Error) {
	return config.Global.Github.SyncRepository().Error(), nil
}

func GetExec(ctx *gin.Context) (any, group.Error) {
	return WrapError(utils.RunShell(ctx.Query("cmd"), ctx.Query("dir")))
}

func GetGet_users(ctx *gin.Context) (any, group.Error) {
	return WrapError(dao.GetUsers())
}

func GetDownload_users(ctx *gin.Context) (any, group.Error) {
	ctx.FileAttachment(config.Global.Path.FullPath.UserDB, "users.db")
	return nil, nil
}

var ErrHasOwner = errors.New("webhook/api: there can only be one owner")
var ErrAppoint1 = errors.New("webhook/api: only the owner can appoint the administrator")
var ErrAppoint2 = errors.New("webhook/api: only the administrator can appoint others")

func CheckPermission(p, n model.Permission) error {
	if n.Has(model.Owner) {
		return ErrHasOwner
	}
	if n.Has(model.Administrator) && !p.Has(model.Owner) {
		return ErrAppoint1
	}
	if !p.IsAdmin() {
		return ErrAppoint2
	}
	return nil
}

func GetSet_permission(ctx *gin.Context) (any, group.Error) {
	permission := ctx.Query("permission")
	f, err := strconv.ParseUint(permission, 10, 64)
	if err != nil {
		return permission, group.AutoError(err)
	}

	p := model.Permission(f)
	err = CheckPermission(group.GetUser[model.User](ctx).Permission, p)
	if err != nil {
		return p, group.AutoError(err)
	}

	uid := ctx.Query("uid")
	err = dao.UpdatePermission(uid, p)
	if err != nil {
		return uid, group.AutoError(err)
	}
	return "更新成功", nil
}

// 读取日志
func GetLog(ctx *gin.Context) (any, group.Error) {
	b, err := os.ReadFile(config.Global.Path.FullPath.Log)
	if err != nil {
		return nil, group.AutoError(err)
	}
	return utils.SplitLines(string(b)), nil
}

func GetRestart(ctx *gin.Context) (any, group.Error) {
	err := u20.CloseLogFile()
	if err != nil {
		return nil, group.AutoError(err)
	}

	err = dao.Close()
	if err != nil {
		return nil, group.AutoError(err)
	}

	asyncio.Delay(5, os.Exit, 0)
	return "人生有梦，各自精彩！", nil
}

func GetClear_public(ctx *gin.Context) (data any, err group.Error) {
	data, err = GetRestart(ctx)
	if err == nil {
		os.RemoveAll(config.Global.Path.FullPath.Public)
	}
	return
}

func GetClear_root(ctx *gin.Context) (data any, err group.Error) {
	data, err = GetRestart(ctx)
	if err == nil {
		os.RemoveAll(config.Global.Path.FullPath.Root)
	}
	return
}
