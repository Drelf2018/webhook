package utils

import (
	"fmt"
	"time"
)

// 重试函数
//
// times: 重试次数 负数则无限制
//
// delay: 休眠秒数 每次重试间休眠时间
//
// f: 要重试的函数 支持格式 func() bool 和 func(T) bool
//
// args: 选填 当函数为后者时会自动将此参数中第一个(args[0])传入
func Retry[T any](times, delay int, f any, args ...T) {
	var do func() bool

	switch f := f.(type) {
	case func() bool:
		do = func() bool { return f() }
	case func(T) bool:
		do = func() bool { return f(args[0]) }
	default:
		panic(fmt.Sprintf("错误的 f: %v(%T)\n", f, f))
	}

	for ; times != 0 && !do(); times-- {
		time.Sleep(time.Duration(delay) * time.Second)
	}
}
