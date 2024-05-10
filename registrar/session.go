package registrar

import (
	"github.com/Drelf2018/request"
)

type Data interface {
	Method() string
	URL() string
}

type Session[T Data] request.Job

func (s *Session[T]) Do() (data T, err error) {
	resp := (*request.Job)(s).Do()
	if err = resp.Error(); err != nil {
		return
	}
	err = resp.Json(&data)
	return
}

func (s *Session[T]) MustDo() (data T) {
	data, _ = s.Do()
	return
}

func NewSession[T Data](opts ...request.Option) *Session[T] {
	var data T
	return (*Session[T])(request.New(data.Method(), data.URL(), opts...))
}
