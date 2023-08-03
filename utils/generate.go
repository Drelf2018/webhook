package utils

import (
	"math"
	"math/rand"
	"strconv"
)

// 随机生成指定位数数字
func RandomNumber(digit int) int {
	if digit <= 0 {
		digit = 6
	}
	// 首位 1-9 其余在 0-10**digit 生成
	// 再将首位乘 10**digit 相加就能保证首位非零
	// 从而保证位数正确
	size := int(math.Pow10(digit - 1))
	return (rand.Intn(9)+1)*size + rand.Intn(size)
}

// 随机生成指定位数数字字符串
func RandomString(digit int) string {
	return strconv.Itoa(RandomNumber(digit))
}
