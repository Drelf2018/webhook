package registrar

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Drelf2018/request"
	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/interfaces"
	"github.com/Drelf2018/webhook/utils"
	u20 "github.com/Drelf2020/utils"
	"github.com/gin-gonic/gin"
)

var (
	KeyOid    string = "__bili_oid__"
	KeyTokens string = "__bili_tokens__"
	QueryUID  string = "uid"
)

var (
	ErrNoToken     = errors.New("webhook/bili: no verification code")
	ErrVerify      = errors.New("webhook/bili: verification failure")
	ErrNonPersonal = errors.New("webhook/bili: non-personal operation")
)

func init() {
	interfaces.SetRegistrar(&BiliRegistrar{
		tokens: make(map[string]string),
	})
}

type BiliRegistrar struct {
	oid     string
	tokens  map[string]string
	session *Session[ApiData]
}

func (b *BiliRegistrar) Initial(c *config.Config) error {
	var ok bool
	b.oid, ok = c.Extra[KeyOid]
	if !ok {
		c.Extra[KeyOid] = "643451139714449427"
		c.Export()
		b.oid = "643451139714449427"
	}

	if tokens, ok := c.Extra[KeyTokens]; ok {
		err := json.Unmarshal([]byte(tokens), &b.tokens)
		if err != nil {
			return err
		}
	}
	if b.tokens == nil {
		b.tokens = make(map[string]string)
	}

	b.session = NewSession[ApiData](request.M{
		"type": "17",
		"oid":  b.oid,
	}.Query)

	return nil
}

func (b *BiliRegistrar) Token(ctx *gin.Context) (data any, err error) {
	uid := ctx.Query(QueryUID)
	if !u20.IsDigit(uid) {
		return nil, fmt.Errorf("webhook/bili: invalid uid: %s", uid)
	}

	token := utils.RandomNumberMixString(6, 3)
	b.tokens[uid] = ctx.RemoteIP() + "_" + token
	return map[string]string{"token": token, "oid": b.oid}, nil
}

func (b *BiliRegistrar) Register(ctx *gin.Context) (uid string, err error) {
	uid = ctx.Query(QueryUID)

	token, ok := b.tokens[uid]
	if !ok {
		err = ErrNoToken
		return
	}

	ip := ctx.RemoteIP()
	if strings.HasPrefix(token, ip) {
		token = strings.TrimPrefix(token, ip+"_")
	} else {
		err = ErrNonPersonal
		return
	}

	var data ApiData
	data, err = b.session.Do()
	if err != nil {
		return
	}
	if data.Code != 0 {
		err = fmt.Errorf("webhook/bili: api returns an error message: %s", data.Message)
		return
	}

	for _, r := range data.Data.Replies {
		if r.Member.Mid == uid && r.Content.Message == token {
			delete(b.tokens, uid)
			return
		}
	}

	err = ErrVerify
	return
}
