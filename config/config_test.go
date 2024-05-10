package config_test

import (
	"testing"

	"github.com/Drelf2018/initial"
	"github.com/Drelf2018/webhook/config"
)

func TestConfig(t *testing.T) {
	c := &config.Config{}
	err := initial.Default(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Try to clone %d times", c.Github.Done())
}
