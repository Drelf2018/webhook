package config_test

import (
	"testing"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/config"
)

func TestConfig(t *testing.T) {
	c := &config.Config{}
	err := initial.Initial(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(c)
	err = c.Export()
	if err != nil {
		t.Fatal(err)
	}
}
