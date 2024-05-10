package registrar

import "net/http"

type Replie struct {
	Member struct {
		Mid   string `json:"mid"`
		Uname string `json:"uname"`
	} `json:"member"`
	Content struct {
		Message string `json:"message"`
	} `json:"content"`
}

type ApiData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Replies []Replie `json:"replies"`
	} `json:"data"`
}

func (ApiData) Method() string {
	return http.MethodGet
}

func (ApiData) URL() string {
	return "https://api.bilibili.com/x/v2/reply"
}
