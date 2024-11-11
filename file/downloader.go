package file

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

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.37"

type Downloader struct {
	Root    string
	clients map[string]Client
}

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

func (d *Downloader) Open(name string) (http.File, error) {
	// 直接打开本地文件
	if !strings.HasPrefix(name, "/http") {
		return http.Dir(d.Root).Open(name)
	}
	// 分离协议和路径
	protocol, path, found := strings.Cut(name[1:], "/")
	if !found {
		return nil, ErrInvalidURL
	}
	// 创建资源文件夹
	fullpath := filepath.Join(d.Root, path)
	_, err := os.Stat(fullpath)
	if err == nil {
		return os.Open(fullpath)
	}
	err = os.MkdirAll(filepath.Dir(fullpath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	// 解析网址
	url := fmt.Sprintf("%s://%s", protocol, path)
	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}
	// 确定下载器
	var (
		client Client
		ok     bool
	)
	for after, found := u.Host, true; !ok && found; _, after, found = strings.Cut(after, ".") {
		client, ok = d.clients[after]
	}
	if !ok {
		client = cli
	}
	// 获取资源
	resp, err := client.Get(url)
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

var _ http.FileSystem = (*Downloader)(nil)

func NewDownloader(root string, client ...Client) *Downloader {
	d := &Downloader{Root: root}
	d.Register(client...)
	return d
}
