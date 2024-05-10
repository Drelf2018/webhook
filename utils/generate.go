package utils

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

var NowRand = rand.New(rand.NewSource(time.Now().Unix()))

func Intn(n int) int {
	return NowRand.Intn(n)
}

// 随机生成指定位数数字
func RandomNumber(digit int) int {
	if digit <= 0 {
		digit = 6
	}
	// 首位 1-9 其余在 0 - 10**digit 生成
	// 再将首位乘 10**digit 相加就能保证首位非零
	// 从而保证位数正确
	size := int(math.Pow10(digit - 1))
	return (Intn(9)+1)*size + Intn(size)
}

// 随机生成指定位数数字字符串
func RandomString(digit int) string {
	return strconv.Itoa(RandomNumber(digit))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// 随机字母
//
// 参考: https://xie.infoq.cn/article/f274571178f1bbe6ff8d974f3
func RandomLetter(n int) []rune {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[Intn(len(letters))]
	}
	return b
}

// 随机混合字母的数字字符串
//
// digit 位数 mix 混入字母个数
func RandomNumberMixString(digit, mix int) string {
	if mix > digit {
		panic("混入比原长度还长你是挺牛逼的")
	}

	isMixed := make([]rune, digit)
	for _, r := range RandomLetter(mix) {
		p := -1
		for p == -1 || isMixed[p] != 0 {
			p = Intn(digit)
		}
		isMixed[p] = r
	}

	for i, r := range RandomString(digit) {
		if isMixed[i] == 0 {
			isMixed[i] = r
		}
	}

	return string(isMixed)
}
