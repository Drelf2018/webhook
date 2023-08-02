package webhook

import (
	"time"
)

type timer struct {
	Server int64            `json:"server"`
	Users  map[string]int64 `json:"users"`
}

var Timer = timer{0, make(map[string]int64)}

// 更新时间戳
func (t *timer) Update(users ...string) *timer {
	stamp := time.Now().Unix()
	t.Server = stamp
	for _, u := range users {
		t.Users[u] = stamp
	}
	return t
}
