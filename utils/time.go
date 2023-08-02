package utils

import (
	"strconv"
	"time"
)

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
		return time.Unix(t.Stamp, 0)
	}
	return time.Time{}
}

func NewTime(t any) Time {
	switch t := t.(type) {
	case string:
		return Time{String: t}
	case int64:
		return Time{Stamp: t}
	case time.Time:
		return Time{Date: t}
	}
	return Time{}
}
