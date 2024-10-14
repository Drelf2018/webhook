package serializer_test

import (
	"testing"

	"github.com/Drelf2018/webhook/model/serializer"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	schema.RegisterSerializer("error", serializer.ErrorSerializer)
}

type RequestLog struct {
	Result []byte
	Error  error `gorm:"serializer:error"`
}

type myError string

func (m myError) Error() string {
	return string(m)
}

func TestError(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		t.Fatal(err)
	}

	gormDB.AutoMigrate(&RequestLog{})

	tx := gormDB.Create(RequestLog{
		Result: []byte(`{"code":0}`),
		Error:  nil,
	})
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}

	tx = gormDB.Create(RequestLog{
		Result: []byte(`{"code":1}`),
		Error:  myError("myError"),
	})
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}

	var logs []RequestLog
	tx = gormDB.Find(&logs)
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}

	for _, log := range logs {
		t.Logf("result: %s, error: %v", log.Result, log.Error)
	}
}
