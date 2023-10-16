package user

import (
	"fmt"
	"net/http"

	"github.com/Drelf2018/request"
	"golang.org/x/exp/slices"
)

type ApiData struct {
	Code int `json:"code"`
	Data struct {
		Replies []Replie `json:"replies"`
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

var api *request.Job

func SetApi(oid string) {
	api = request.New(
		http.MethodGet,
		"https://api.bilibili.com/x/v2/reply",
		request.Datas(request.M{"pn": "1", "type": "17", "oid": oid, "sort": "0"}),
	)
}

// 返回最近回复
func GetReplies() ([]Replie, error) {
	resp := api.Request()
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	var Api ApiData
	err := resp.Json(&Api)
	if err != nil {
		return nil, err
	}
	if Api.Code != 0 {
		return nil, fmt.Errorf("api error code: %v", Api.Code)
	}
	return Api.Data.Replies, nil
}

// 检查回复
func (u User) MatchReplies() (bool, error) {
	rep, err := GetReplies()
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(rep, func(r Replie) bool {
		return r.Member.Mid == u.Uid && r.Content.Message == u.Token
	}), nil
}
