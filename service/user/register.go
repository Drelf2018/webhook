package user

import (
	"github.com/Drelf2018/webhook/utils"
	uuid "github.com/satori/go.uuid"
)

var (
	Tokens = make(map[string]User)
	Uids   = make(map[string]string)
)

// 获取随机 Token
func GetRandomToken(uid string) (auth, token string) {
	if t, ok := Uids[uid]; ok {
		delete(Tokens, t)
	}
	token = utils.RandomNumberMixString(6, 3)
	auth = uuid.NewV4().String()
	Tokens[auth] = User{Uid: uid, Token: token}
	Uids[uid] = auth
	return
}

func Get(auth string) User {
	return Tokens[auth]
}

// 清理
func Done(uid string) {
	if t, ok := Uids[uid]; ok {
		delete(Tokens, t)
	}
	delete(Uids, uid)
}
