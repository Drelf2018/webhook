package utils

import "sync"

// 异步运行 tasks 中的每一个函数
func All(tasks ...func()) {
	wg := sync.WaitGroup{}
	wg.Add(len(tasks))
	do := func(f func()) {
		defer wg.Done()
		f()
	}
	for _, f := range tasks {
		go do(f)
	}
	wg.Wait()
}

// 异步运行同一函数多次 参数为运行序号
func List(task func(i int), length int) {
	wg := sync.WaitGroup{}
	wg.Add(length)
	do := func(i int) {
		defer wg.Done()
		task(i)
	}
	for i := 0; i < length; i++ {
		go do(i)
	}
	wg.Wait()
}

// 异步运行同一函数多次 参数用户提供 返回参数顺序对应的结果
func Await[T any, V any, A ~[]T](task func(T) V, args *A) []V {
	length := len(*args)
	wg := sync.WaitGroup{}
	wg.Add(length)
	result := make([]V, length)
	do := func(i int, arg T) {
		defer wg.Done()
		result[i] = task(arg)
	}
	for i, arg := range *args {
		go do(i, arg)
	}
	wg.Wait()
	return result
}

// 回调函数
func Callback[T any](f func(...any) T, callback func(T, ...any), args ...any) {
	callback(f(args), args)
}

type EventLoop[T any] struct {
	sync.WaitGroup
}

func (el *EventLoop[T]) AddF(f func()) {
	el.Add(1)
	go func() {
		f()
		el.Done()
	}()
}

func (el *EventLoop[T]) AddFunc(f func() T) {
	el.Add(1)
	go func() {
		f()
		el.Done()
	}()
}

func (el *EventLoop[T]) AddTask(f func(...any) T, args ...any) {
	el.Add(1)
	go func() {
		f(args)
		el.Done()
	}()
}
