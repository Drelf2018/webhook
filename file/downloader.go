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

type Downloader struct {
	Root    string
	Options map[string]string
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
	var (
		client Client
		ok     bool
	)
	if len(d.clients) != 0 {
		for after, found := u.Host, true; !ok && found; _, after, found = strings.Cut(after, ".") {
			client, ok = d.clients[after]
		}
	}
	// 获取资源
	var resp *http.Response
	if ok {
		resp, err = client.Get(url, d.Options["/"+protocol+"/"+u.Host+u.Path])
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

func (d *Downloader) Download(url string) (http.File, error) {
	return d.Open("/" + strings.Replace(url, ":/", "", 1))
}

var _ http.FileSystem = (*Downloader)(nil)

func NewDownloader(root string, client ...Client) *Downloader {
	d := &Downloader{Root: root, Options: make(map[string]string)}
	d.Register(client...)
	return d
}
