package user

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// 监听列表获取错误
var ErrNotListeningList = errors.New("不是一个好的监听列表")

// 监听列表的读取实现
type Listening []string

func (l *Listening) Scan(val any) error {
	if val, ok := val.(string); ok {
		if val == "" {
			*l = make(Listening, 0)
			return nil
		}
		*l = strings.Split(val, ",")
		return nil
	}
	return ErrNotListeningList
}

func (l Listening) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}
