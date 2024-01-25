package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/Drelf2018/webhook/model"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type Api struct {
	Visitor
	Submitter
	Admin
}

// 解决跨域问题
//
// 参考: https://blog.csdn.net/u011866450/article/details/126958238
func Cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	// 禁止所有 OPTIONS 方法 原因见博文
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func (Api) Use(r *gin.RouterGroup) {
	r.Use(Cors)
	r.Use(static.ServeRoot("/", "./views"))
	r.Use(static.ServeRoot("/user", "./views"))
}

type fileSystem string

func (fs fileSystem) Open(name string) (http.File, error) {
	if !strings.HasPrefix(name, "/http") {
		return http.Dir(fs).Open(name)
	}
	if _, err := os.Stat(model.ConvertURL(name)); err != nil {
		model.Download(name[1:])
	}
	return http.Dir(fs).Open(model.CleanURL(name))
}

func (Api) StaticFSPublic() (string, http.FileSystem) {
	return "/public", fileSystem("public")
}
