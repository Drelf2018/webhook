package utils

type void struct{}

// 集合
//
// 参考: https://www.zhihu.com/question/582450146
type Set[T comparable] map[T]void

// 添加
func (s *Set[T]) Add(value T) {
	(*s)[value] = void{}
}

// 移除
func (s *Set[T]) Remove(value T) {
	delete(*s, value)
}

// 存在
func (s *Set[T]) Contains(value T) bool {
	_, ok := (*s)[value]
	return ok
}

// 占用
func (s *Set[T]) Size() int {
	return len(*s)
}
