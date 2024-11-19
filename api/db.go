package api

import (
	"errors"
	"strings"

	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var userDB, blogDB *gorm.DB

func UserDB() *gorm.DB {
	if userDB == nil {
		var err error
		userDB, err = gorm.Open(sqlite.Open(webhook.Global().Path.Full.UserDB))
		if err != nil {
			panic(err)
		}
		err = userDB.AutoMigrate(&model.User{}, &model.Task{}, &model.Filter{}, &model.RequestLog{})
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

func CloseDB() error {
	var errs []string

	if userDB, err := UserDB().DB(); err != nil {
		errs = append(errs, err.Error())
	} else if err = userDB.Close(); err != nil {
		errs = append(errs, err.Error())
	}

	if blogDB, err := BlogDB().DB(); err != nil {
		errs = append(errs, err.Error())
	} else if err = blogDB.Close(); err != nil {
		errs = append(errs, err.Error())
	}

	return errors.New(strings.Join(errs, "; "))
}
