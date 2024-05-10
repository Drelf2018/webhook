package model_test

import (
	"testing"

	"github.com/Drelf2018/webhook/database/model"
)

func TestSend(t *testing.T) {
	resp := model.TestJob.Send(model.TestPost)
	err := resp.Error()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resp.Text(): %v\n", resp.Text())
}
