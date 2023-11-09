package utils

import (
	"strconv"
	"time"
)

var local = time.FixedZone("CST", 8*3600)

// 时间类 方便相互转换
type Time struct {
	String string
	Stamp  int64
	Date   time.Time
}

func (t Time) ToString() string {
	if t.String != "" {
		return t.String
	}
	if !t.Date.IsZero() {
		t.Stamp = t.Date.Unix()
	}
	if t.Stamp != 0 {
		return strconv.Itoa(int(t.Stamp))
	}
	return ""
}

func (t Time) ToStamp() int64 {
	if t.Stamp != 0 {
		return t.Stamp
	}
	if !t.Date.IsZero() {
		return t.Date.Unix()
	}
	if t.String != "" {
		stamp, err := strconv.ParseInt(t.String, 10, 64)
		if err == nil {
			return stamp
		}
	}
	return 0
}

func (t Time) ToDate() time.Time {
	if !t.Date.IsZero() {
		return t.Date
	}
	if t.String != "" {
		t.Stamp, _ = strconv.ParseInt(t.String, 10, 64)
	}
	if t.Stamp != 0 {
		return time.Unix(t.Stamp, 0).In(local)
	}
	return time.Time{}
}

// 以当前为基准延后
func (t Time) Delay(seconds int64) Time {
	return Time{Stamp: t.ToStamp() + seconds}
}

// 支持格式 string int64 time.Time
//
// 特别的 使用 nil 时将以当前时间创建
func NewTime(t any) Time {
	switch t := t.(type) {
	case string:
		return Time{String: t}
	case int64:
		return Time{Stamp: t}
	case time.Time:
		return Time{Date: t}
	case nil:
		return Time{Stamp: time.Now().Unix()}
	}
	return Time{}
}

var timer = struct {
	Server int64            `json:"server"`
	Users  map[string]int64 `json:"users"`
}{
	Server: time.Now().Unix(),
	Users:  make(map[string]int64),
}

// 更新时间戳
func Timer(uids ...string) any {
	timer.Server = time.Now().Unix()
	for _, uid := range uids {
		timer.Users[uid] = timer.Server
	}
	return timer
}
