package utils

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slices"

	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/config"
)

type ApiData struct {
	Code int `json:"code"`
	Data struct {
		Replies []*Replie `json:"replies"`
	} `json:"data"`
}

type Replie struct {
	Member struct {
		Mid   string `json:"mid"`
		Uname string `json:"uname"`
	} `json:"member"`
	Content struct {
		Message string `json:"message"`
	} `json:"content"`
}

func Get[T any](url string, params request.M) (response T) {
	(&request.Job{Method: http.MethodGet, Url: url, Data: params}).Request().Json(&response)
	return
}

func SearchToken(uid, token string) (bool, error) {
	api := Get[ApiData]("https://api.bilibili.com/x/v2/reply", request.M{
		"type": "17",
		"oid":  config.Global.Oid,
	})
	if api.Code != 0 {
		return false, fmt.Errorf("api error code: %v", api.Code)
	}
	return slices.ContainsFunc(api.Data.Replies, func(r *Replie) bool {
		return r.Member.Mid == uid && r.Content.Message == token
	}), nil
}
