package user

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Drelf2020/utils/request"
)

var Url string

// 构建网址
func MakeUrl(OID string) {
	Url = fmt.Sprintf("https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies", OID)
}

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
	var Api ApiData
	err := request.Get(Url).Json(&Api)
	if err != nil {
		return nil, err
	}
	if Api.Code != 0 {
		return nil, errors.New("返回代码：" + strconv.Itoa(Api.Code) + " 错误")
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
