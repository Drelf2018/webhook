package registrar_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/Drelf2018/webhook/config"
	"github.com/Drelf2018/webhook/registrar"
	"github.com/gin-gonic/gin"
)

func TestBili(t *testing.T) {
	var reg registrar.BiliRegistrar

	err := reg.Initial(&config.Config{Extra: map[string]string{
		registrar.KeyOid:    "643451139714449427",
		registrar.KeyTokens: `{"188888131":"127.0.0.1_488KPX","senpai":"1.145.14.191_9810"}`,
	}})
	if err != nil {
		t.Fatal(err)
	}

	c := &gin.Context{Request: &http.Request{
		URL:        &url.URL{RawQuery: "uid=188888131"},
		RemoteAddr: "127.0.0.1:8080",
	}}

	uid, err := reg.Register(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(uid)

	c.Request.RemoteAddr = "192.168.0.1:9000"
	_, err = reg.Register(c)
	if err != nil {
		t.Log(err)
	}

	c = &gin.Context{Request: &http.Request{
		URL:        &url.URL{RawQuery: "uid=senpai"},
		RemoteAddr: "1.145.14.191:9810",
	}}
	_, err = reg.Register(c)
	if err != nil {
		t.Log(err)
	}
}
