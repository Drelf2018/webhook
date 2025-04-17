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

// 当前版本信息
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

// 获取当前在线状态
func GetOnline(ctx *gin.Context) (any, error) {
	now := time.Now()
	m := make(map[string]int64)
	onlineUsers.Range(func(key, value any) bool {
		m[key.(string)] = now.Sub(value.(time.Time)).Milliseconds()
		return true
	})
	return m, nil
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
