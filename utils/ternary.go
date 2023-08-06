package utils

// 三目运算符
func Ternary[T any](expr bool, a, b T) T {
	if expr {
		return a
	}
	return b
}
