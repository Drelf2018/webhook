package network

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Drelf2020/utils/request"
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

var (
	OID = ""
	Url = ""
)

// 构建网址
func MakeUrl() {
	Url = fmt.Sprintf("https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies", OID)
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
func MatchReplies(uid, pwd string) (bool, error) {
	rs, err := GetReplies()
	if err != nil {
		return false, err
	}
	for _, r := range rs {
		if r.Member.Mid != uid {
			continue
		}
		if r.Content.Message == pwd {
			return true, nil
		}
	}
	return false, nil
}
