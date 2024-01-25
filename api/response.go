package api

import (
	"errors"
	"net/http"
	"runtime"

	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
}

var _ render.Render = &Response{}
var ErrOddParameter = errors.New("only an even number of arguments can be passed in")

// 解析数据
func (r *Response) Set(data ...any) {
	l := len(data)
	switch {
	case l == 0:
		return
	case l == 1:
		r.Data = data[0]
	case l&1 == 1:
		panic(ErrOddParameter)
	default:
		m := make(map[any]any)
		for i := 0; i < l; i += 2 {
			m[data[i]] = data[i+1]
		}
		r.Data = m
	}
}

func (r *Response) Render(w http.ResponseWriter) error {
	return render.WriteJSON(w, r)
}

func (r *Response) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

func resp(key any, data ...any) (r *Response) {
	r = &Response{}
	switch v := key.(type) {
	case nil:
		r.Code = 0
	case int:
		r.Code = v
	case string:
		_, _, r.Code, _ = runtime.Caller(2)
		r.Message = v
	case error:
		_, _, r.Code, _ = runtime.Caller(2)
		r.Message = v.Error()
	}
	r.Set(data...)
	return
}

func Abort(c *gin.Context, key any, data ...any) {
	c.Abort()
	c.Render(200, resp(key, data...))
}

func Succeed(c *gin.Context, data ...any) {
	c.Render(200, resp(nil, data...))
}

func User(c *gin.Context) *model.User {
	return c.MustGet("user").(*model.User)
}
