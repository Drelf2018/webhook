package webhook

import (
	"path/filepath"
	"strconv"

	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 设置默认值
func Default[T comparable](a *T, b T) {
	var zero T
	if *a == zero {
		*a = b
	}
}

// 资源
type Resource struct {
	// 文件夹路径
	Path string
	// 公开子文件夹名
	Public string
	// 数据库名
	Database string
}

func (r *Resource) fill() {
	Default(&r.Path, "resource")
	Default(&r.Public, "public")
	Default(&r.Database, "posts.db")
}

func (r Resource) ToRoot(files ...string) string {
	elem := make([]string, len(files)+1)
	elem[0] = r.Path
	for i, file := range files {
		elem[i+1] = file
	}
	// return filepath.Join(r.Path, files...)
	// 甚至不允许这样 这语言是有够傻逼的
	return filepath.Join(elem...)
}

func (r Resource) ToPublic() string {
	return filepath.Join(r.Path, r.Public)
}

func (r Resource) ToPostsDB() gorm.Dialector {
	dsn := filepath.Join(r.Path, r.Public, r.Database)
	return sqlite.Open(dsn)
}

func (r Resource) ToUsersDB() gorm.Dialector {
	dsn := filepath.Join(r.Path, "users.db")
	return sqlite.Open(dsn)
}

// 网络
type Network struct {
	// 服务器启动 IP
	Url string
	// 启动端口
	Port int
}

func (n *Network) fill() {
	Default(&n.Url, "0.0.0.0")
	Default(&n.Port, 9000)
}

func (n Network) Addr() string {
	return n.Url + ":" + strconv.Itoa(n.Port)
}

// 运行环境
type Runtime struct {
	// 自定义全接口
	//
	// DIY 不为 nil 时仅执行此函数 不会执行下面的鉴权前后函数
	DIY func(r *Config)
	// 鉴权前
	BeforeAuthorize func(r *Config)
	// 鉴权后
	AfterAuthorize func(r *Config)
}

func (r *Runtime) fill() {
	if r.BeforeAuthorize == nil {
		r.BeforeAuthorize = BeforeAuthorize
	}
	if r.AfterAuthorize == nil {
		r.AfterAuthorize = AfterAuthorize
	}
}

// webhook 配置
type Config struct {
	// gin 启动模式
	Debug bool
	// 资源
	Resource
	// 网络
	Network
	// 运行
	Runtime
	// 主页 git 链接 只需填写前三项
	Github
	// 其他参数
	Map gin.H
	// 引擎
	*gin.Engine
}

// 自动填充
func (r *Config) fill() {
	gin.SetMode(utils.Ternary(r.Debug, gin.DebugMode, gin.ReleaseMode))
	r.Resource.fill()
	r.Network.fill()
	r.Runtime.fill()
	r.Github.fill(&r.Resource)
	if r.Engine == nil {
		r.Engine = gin.Default()
	}
}
