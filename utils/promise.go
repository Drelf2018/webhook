package utils

import (
	"fmt"
	"sync"
	"time"
)

// 事件循环
type EventLoop[T any, V any, R ~[]V] struct {
	sync.WaitGroup
	Results *R
}

// 添加任务
func (el *EventLoop[T, V, R]) AddTask(f any, args ...T) {
	el.Add(1)
	position := len(*el.Results)
	*el.Results = append(*el.Results, *new(V))
	go func(p int) {
		switch f := f.(type) {
		case func(T):
			f(args[0])
		case func(T) V:
			(*el.Results)[p] = f(args[0])
		case func():
			f()
		case func() V:
			(*el.Results)[p] = f()
		default:
			panic(fmt.Sprintf("错误的 f: %v(%T)\n", f, f))
		}
		el.Done()
	}(position)
}

// 异步运行 tasks 中的每一个函数
func All(tasks ...func()) {
	loop := EventLoop[any, any, []any]{Results: &[]any{}}
	for _, f := range tasks {
		loop.AddTask(f)
	}
	loop.Wait()
}

// 异步运行同一函数多次 参数为运行序号
func List(task func(i int), length int) {
	loop := EventLoop[int, any, []any]{Results: &[]any{}}
	for i := 0; i < length; i++ {
		loop.AddTask(task, i)
	}
	loop.Wait()
}

// 异步运行同一函数多次 参数用户提供 返回参数顺序对应的结果
func Await[T any, V any, A ~[]T](task func(T) V, args *A) []V {
	loop := EventLoop[T, V, []V]{Results: &[]V{}}
	for _, arg := range *args {
		loop.AddTask(task, arg)
	}
	loop.Wait()
	return *loop.Results
}

// 异步运行同一函数多次 参数用户提供 返回参数顺序对应的结果
func AwaitWith[T any, V any, A ~[]T, R ~[]V](task func(T) V, args *A, results *R) {
	loop := EventLoop[T, V, R]{Results: results}
	for _, arg := range *args {
		loop.AddTask(task, arg)
	}
	loop.Wait()
}

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
