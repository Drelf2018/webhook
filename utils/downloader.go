package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"
)

// 客户端接口
type Client interface {
	// 在选择客户端时，会根据主机名进行匹配
	// 返回值包含 example.com 时会匹配所有 *.example.com 的请求
	Hosts() []string

	// 自定义请求方法，可以添加请求头、鉴权参数、Cookies 之类
	Get(url string) (*http.Response, error)
}

// 下载器
//
// 对于一个地址为 https://www.example.com/public/data.txt 文件
// 会保存在 $Root/www.example.com/public/data.txt
type Downloader struct {
	// 保存文件的根目录
	Root string

	// 使用的自定义客户端
	clients map[string]Client
}

// 注册客户端
func (d *Downloader) Register(client ...Client) {
	if d.clients == nil {
		d.clients = make(map[string]Client)
	}
	for _, c := range client {
		for _, host := range c.Hosts() {
			d.clients[host] = c
		}
	}
}

var ErrInvalidURL = errors.New("webhook/file: invalid URL")

// 打开文件
//
// 会先尝试在本地访问，如果不存在再从网络下
func (d *Downloader) Open(name string) (http.File, error) {
	// 直接打开本地文件
	if !strings.HasPrefix(name, "/http") {
		return http.Dir(d.Root).Open(strings.ReplaceAll(name, ":", "/"))
	}
	// 分离协议和路径
	protocol, path, found := strings.Cut(name[1:], "/")
	if !found {
		return nil, ErrInvalidURL
	}
	// 解析网址
	url := fmt.Sprintf("%s://%s", protocol, path)
	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Hostname() == "" {
		return nil, ErrInvalidURL
	}
	// 创建资源文件夹
	fullpath := filepath.Join(d.Root, u.Hostname(), u.Port(), u.Path)
	_, err = os.Stat(fullpath)
	if err == nil {
		return os.Open(fullpath)
	}
	err = os.MkdirAll(filepath.Dir(fullpath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	// 确定下载器
	var client Client
	if len(d.clients) != 0 {
		for after, found := u.Host, true; client == nil && found; _, after, found = strings.Cut(after, ".") {
			client = d.clients[after]
		}
	}
	// 获取资源
	var resp *http.Response
	if client != nil {
		resp, err = client.Get(url)
	} else {
		resp, err = http.Get(url)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 保存并打开文件
	err = os.WriteFile(fullpath, content, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Open(fullpath)
}

// 下载文件
func (d *Downloader) Download(url string) (http.File, error) {
	return d.Open("/" + strings.Replace(url, ":/", "", 1))
}

var _ http.FileSystem = (*Downloader)(nil)

func NewDownloader(root string, client ...Client) *Downloader {
	d := &Downloader{Root: root}
	d.Register(client...)
	return d
}
