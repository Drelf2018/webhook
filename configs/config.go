package configs

import (
	"github.com/Drelf2018/resource"
	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

// 设置值类型对象默认值
func SetZero[T comparable](a *T, b T) {
	var zero T
	if *a == zero {
		*a = b
	}
}

// 设置引用类型对象默认值
func SetNil[C any, T utils.CanNil[C]](a *T, b T) {
	if *a == nil {
		*a = b
	}
}

// 资源文件夹路径
type Path struct {
	Root   string
	Views  string
	Public string
	Posts  string
	Users  string
}

func (p *Path) Init() {
	SetZero(&p.Root, "resource")
	SetZero(&p.Views, "views")
	SetZero(&p.Public, "public")
	SetZero(&p.Posts, "posts.db")
	SetZero(&p.Users, "users.db")
}

// webhook 配置
type Config struct {
	// 引擎
	*gin.Engine
	// 端口 0~65535
	Port uint16
	// 资源文件管理器
	Resource resource.Explorer
	// 资源文件夹路径
	Path Path
	// Github 主页
	Github Github
	// 管理员
	Administrators []string
}

// 自动填充
func (r *Config) Init() *Config {
	r.Path.Init()
	SetZero(&r.Port, 9000)
	SetNil[gin.Engine](&r.Engine, gin.Default())
	SetNil[string](&r.Administrators, []string{})

	if r.Resource == nil {
		r.Resource = new(resource.Resource).Init(r.Path.Root).Root
	}
	data.SetPublic("/"+r.Path.Public, r.Resource.MakeTo(r.Path.Public))

	r.Github.Init()

	return r
}

// 更新主页
func (r *Config) UpdateIndex() (err error) {
	// 先获取最新版本
	views := r.Resource.MakeTo(r.Path.Views)
	views.MkdirAll()

	err = r.Github.GetLatestCommit()
	if err != nil {
		return
	}

	if views.Find("index.html") != nil {
		if ver := views.Find(".version"); ver != nil {
			if ver.MustRead() == r.Github.Commit.Sha {
				return nil
			}
		}
	}

	// 再决定要不要克隆
	folder := views.Path()
	views.RemoveAll()
	err = r.Github.Clone(folder)
	if err != nil {
		return
	}

	ver, _ := r.Resource.MakeTo(r.Path.Views).Touch(".version", 0)
	return ver.Write(r.Github.Commit.Sha)
}

type LifeCycle interface {
	// 初始化
	OnCreate(*Config)
	// 跨域设置
	OnCors(*Config)
	// 静态资源绑定
	OnStatic(*Config)
	// 访客接口
	Visitor(*Config)
	// 鉴定提交者权限
	OnAuthorize(*Config)
	// 提交者接口
	Submitter(*Config)
	// 鉴定管理员权限
	OnAdmin(*Config)
	// 管理员接口
	Administrator(*Config)
	// 绑定所有接口
	Bind(*Config)
}
