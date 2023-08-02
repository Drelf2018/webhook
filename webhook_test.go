package webhook_test

import (
	"testing"

	"github.com/Drelf2018/webhook"
)

func TestMain(t *testing.T) {
	webhook.Run(&webhook.Config{Debug: true})
}
