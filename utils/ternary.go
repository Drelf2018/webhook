package utils

// 三目运算符
func Ternary[T any](expr bool, a, b T) T {
	if expr {
		return a
	}
	return b
}

func Default[T comparable](a *T, b T) {
	var zero T
	if *a == zero {
		*a = b
	}
}
