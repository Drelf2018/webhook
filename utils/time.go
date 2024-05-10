package utils

import (
	"encoding/json"
	"time"

	"github.com/Drelf2018/intime"
)

type onlineList map[string]intime.Time

func (o onlineList) MarshalJSON() ([]byte, error) {
	now := time.Now().UnixMilli()
	m := make(map[string]int64)
	for uid, t := range o {
		m[uid] = now - t.UnixMilli()
	}
	return json.Marshal(m)
}

func (o onlineList) Update(uid string) {
	o[uid] = intime.Now()
}

var OnlineList = make(onlineList)
