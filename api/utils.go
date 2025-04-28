package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

const (
	Version string = "v0.19.0"
	Success string = "success"
)

var version = struct {
	Api   string    `json:"api"`
	Env   string    `json:"env"`
	Start time.Time `json:"start"`
	Index []string  `json:"index"`
}{
	Api: Version,
	Env: fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
}

func init() {
	time.Local, _ = time.LoadLocation("Asia/Shanghai")
	version.Start = time.Now()
}

// 获取版本信息
func GetVersion(ctx *gin.Context) (any, error) {
	return version, nil
}

// 校验鉴权码
func GetValid(ctx *gin.Context) (any, error) {
	_, err := JWTAuth(ctx)
	return err == nil, nil
}

var onlineUsers sync.Map // map[string]time.Time

// 更新在线时间
func GetPing(ctx *gin.Context) (any, error) {
	uid, err := JWTAuth(ctx)
	if err != nil {
		Error(ctx, err)
		return 1, err
	}
	onlineUsers.Store(uid, time.Now())
	return "pong", nil
}

// 获取在线状态
func GetOnline(ctx *gin.Context) (any, error) {
	now := time.Now()
	m := make(map[string]int64)
	onlineUsers.Range(func(key, value any) bool {
		m[key.(string)] = now.Sub(value.(time.Time)).Milliseconds()
		return true
	})
	return m, nil
}

// 获取 Token
func GetToken(ctx *gin.Context) (data any, err error) {
	// 先从请求参数获取账号信息
	var q struct {
		UID     string `form:"uid"`
		PWD     string `form:"pwd"`
		Refresh bool   `form:"refresh"`
	}
	err = ctx.ShouldBindQuery(&q)
	if err != nil {
		return 1, err
	}
	// 否则从请求头获取
	if q.UID == "" {
		q.UID, q.PWD, err = registrar.BasicAuth(ctx)
		if err != nil {
			return 2, err
		}
	}
	// 校验账号存在和密码正确
	user := &model.User{UID: q.UID}
	tx := UserDB.Limit(1).Find(user)
	if tx.Error != nil {
		return 3, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 4, ErrUserNotExist
	}
	if user.Password != q.PWD {
		return 5, ErrIncorrectPwd
	}
	// 校验账号是否封禁
	now := time.Now()
	if user.Ban.After(now) {
		return 6, ErrBanned
	}
	// 使用当前时间创建用户声明
	// 若已有声明且不刷新则沿用
	// 否则将当前时间保存在数据库中
	claim := &UserClaims{user.UID, now.UnixMilli()}
	iat, found := tokenIssuedAt.Load(user.UID)
	if found && !q.Refresh {
		claim.IssuedAt = iat.(int64)
	} else {
		err = UserDB.Updates(claim).Error
		if err != nil {
			return 7, err
		}
		tokenIssuedAt.Store(claim.UID, claim.IssuedAt)
	}
	// 生成 Token
	token, err := claim.Token()
	if err != nil {
		return 8, err
	}
	return token, nil
}

// 请求转发 https://blog.csdn.net/qq_29799655/article/details/113841064
func AnyForwardURL(ctx *gin.Context) (any, error) {
	// 复刻请求
	url := strings.TrimPrefix(ctx.Param("url"), "/")
	req := ctx.Request.Clone(context.Background())
	req.URL.Scheme, url, _ = strings.Cut(url, "/")
	req.URL.Host, url, _ = strings.Cut(url, "/")
	req.URL.Path = "/" + url
	req.Host = req.URL.Host
	req.RemoteAddr = ""
	req.RequestURI = ""
	// 发送新请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1, err
	}
	// 写入状态码和 Header
	ctx.Status(resp.StatusCode)
	header := ctx.Writer.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	// 将 Body 写回原请求中
	_, err = io.Copy(ctx.Writer, resp.Body)
	if err != nil {
		return 2, err
	}
	return nil, nil
}

// 下载资源
func DownloadAssets(blog *model.Blog) error {
	if downloader == nil || blog == nil {
		return nil
	}
	errs := make(utils.JoinError, 0)
	if blog.Avatar != "" {
		_, err := downloader.Download(blog.Avatar)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, url := range blog.Assets {
		_, err := downloader.Download(url)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, url := range blog.Banner {
		_, err := downloader.Download(url)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if blog.Reply != nil {
		err := DownloadAssets(blog.Reply)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
