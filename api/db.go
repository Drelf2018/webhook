package api

import (
	"errors"
	"time"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
	"github.com/Drelf2018/webhook/model/runner"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var userDB, blogDB *gorm.DB
var taskRunner *runner.TaskRunner

func UserDB() *gorm.DB {
	if userDB == nil {
		var err error
		userDB, err = gorm.Open(sqlite.Open(webhook.Global().Path.Full.UserDB))
		if err != nil {
			panic(err)
		}
		err = userDB.AutoMigrate(&model.User{}, &model.Task{}, &model.RequestLog{})
		if err != nil {
			panic(err)
		}
	}
	return userDB
}

func BlogDB() *gorm.DB {
	if blogDB == nil {
		var err error
		blogDB, err = gorm.Open(sqlite.Open(webhook.Global().Path.Full.BlogDB))
		if err != nil {
			panic(err)
		}
		err = blogDB.AutoMigrate(&model.Blog{})
		if err != nil {
			panic(err)
		}
	}
	return blogDB
}

func Runner() *runner.TaskRunner {
	if taskRunner == nil {
		taskRunner = &runner.TaskRunner{TaskDB: UserDB(), Timeout: 10 * time.Second}
	}
	return taskRunner
}

func CloseDB() error {
	var errs []error
	userDB, err := UserDB().DB()
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, userDB.Close())
	}
	blogDB, err := BlogDB().DB()
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, blogDB.Close())
	}

	var s string
	for _, err := range errs {
		if err != nil {
			s += err.Error()
		}
	}
	if s == "" {
		return nil
	}
	return errors.New(s)
}
