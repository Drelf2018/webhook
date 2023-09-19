package user

import (
	"fmt"

	"github.com/Drelf2018/request"
)

type ApiData struct {
	Code int       `json:"code"`
	Data []Replies `json:"data"`
}

type Replies struct {
	Member struct {
		Mid   string `json:"mid"`
		Uname string `json:"uname"`
	} `json:"member"`
	Content struct {
		Message string `json:"message"`
	} `json:"content"`
}

// 返回最近回复
func GetReplies() ([]Replies, error) {
	resp := request.Get(url)
	if resp.Error != nil {
		return nil, resp.Error
	}
	var Api ApiData
	err := resp.Json(&Api)
	if err != nil {
		return nil, err
	}
	if Api.Code != 0 {
		return nil, fmt.Errorf("返回错误代码: %v", Api.Code)
	}
	return Api.Data, nil
}

// 检查回复
func (u User) MatchReplies() (bool, error) {
	rs, err := GetReplies()
	if err != nil {
		return false, err
	}
	for _, r := range rs {
		if r.Member.Mid != u.Uid {
			continue
		}
		if r.Content.Message == u.Token {
			return true, nil
		}
	}
	return false, nil
}
