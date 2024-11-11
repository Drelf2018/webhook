package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

func init() {
	registrar.SetRegistrar(&Reply{})
}

type ReplyPayload struct {
	OID      string `json:"oid"`
	UID      string `json:"uid"`
	Password string `json:"password"`
}

type ReplyData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Replies []struct {
			Member struct {
				Uname string `json:"uname"`
				MID   string `json:"mid"`
			} `json:"member"`
			Content struct {
				Message string `json:"message"`
			} `json:"content"`
		} `json:"replies"`
	} `json:"data"`
}

type Reply struct {
	SessData string
	BiliJct  string
	Buvid3   string

	m sync.Map // map[string][2]string
}

func (r *Reply) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		switch cookie.Name {
		case "SESSDATA":
			r.SessData = cookie.Value
		case "bili_jct":
			r.BiliJct = cookie.Value
		case "buvid3":
			r.Buvid3 = cookie.Value
		}
	}
}

func (r *Reply) Cookies(*url.URL) []*http.Cookie {
	return []*http.Cookie{
		{Name: "SESSDATA", Value: r.SessData},
		{Name: "bili_jct", Value: r.BiliJct},
		{Name: "buvid3", Value: r.Buvid3},
	}
}

var _ http.CookieJar = (*Reply)(nil)

func (r *Reply) Initial(extra map[string]any) error {
	get := func(key string) string {
		v, ok := extra[key]
		if !ok {
			extra[key] = ""
			return ""
		}
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprint(v)
		}
		return s
	}
	r.SessData = get("SESSDATA")
	r.BiliJct = get("bili_jct")
	r.Buvid3 = get("buvid3")
	return nil
}

func (r *Reply) Register(ctx *gin.Context) (user any, data any, err error) {
	var payload ReplyPayload
	err = ctx.ShouldBindJSON(&payload)
	if err != nil {
		return nil, 10001, err
	}

	if payload.OID == "" {
		code := uniuri.New()
		r.m.Store(ctx.RemoteIP(), [2]string{payload.UID, code})
		return nil, code, nil
	}

	client := &http.Client{Jar: r}
	resp, err := client.Get("https://api.bilibili.com/x/v2/reply?type=17&oid=" + payload.OID)
	if err != nil {
		return nil, 10002, err
	}
	defer resp.Body.Close()

	var pdata ReplyData
	err = json.NewDecoder(resp.Body).Decode(&pdata)
	if err != nil {
		return nil, 10003, err
	}
	if pdata.Code != 0 {
		return nil, 10004, fmt.Errorf("webhook/cmd/webhook: \"%s\" with code %d", pdata.Message, pdata.Code)
	}

	v, ok := r.m.Load(ctx.RemoteIP())
	if !ok {
		return nil, 10005, errors.New("webhook/cmd/webhook: no verification code is generated")
	}
	val := v.([2]string)
	for _, reply := range pdata.Data.Replies {
		if reply.Member.MID == val[0] && reply.Content.Message == val[1] {
			r.m.Delete(val[0])
			return &model.User{UID: val[0], Name: reply.Member.Uname, Password: payload.Password}, 0, nil
		}
	}
	return nil, 10006, registrar.ErrVerify
}
