package user

import (
	"fmt"
	"net/http"

	"github.com/Drelf2018/request"
	"golang.org/x/exp/slices"
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

var api *request.Job

func SetApi(oid string) {
	api = request.New(
		http.MethodGet,
		fmt.Sprintf(
			"https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies",
			oid,
		),
	)
}

// 返回最近回复
func GetReplies() ([]Replies, error) {
	resp := api.Request()
	if resp.Error != nil {
		return nil, resp.Error
	}
	var Api ApiData
	err := resp.Json(&Api)
	if err != nil {
		return nil, err
	}
	if Api.Code != 0 {
		return nil, fmt.Errorf("api error code: %v", Api.Code)
	}
	return Api.Data, nil
}

// 检查回复
func (u User) MatchReplies() (bool, error) {
	rep, err := GetReplies()
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(rep, func(r Replies) bool {
		return r.Member.Mid == u.Uid && r.Content.Message == u.Token
	}), nil
}
