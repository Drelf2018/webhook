package api

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var UserDB, BlogDB *gorm.DB

func CloseDB() error {
	var errs []string

	if userDB, err := UserDB.DB(); err != nil {
		errs = append(errs, err.Error())
	} else if err = userDB.Close(); err != nil {
		errs = append(errs, err.Error())
	}

	if blogDB, err := BlogDB.DB(); err != nil {
		errs = append(errs, err.Error())
	} else if err = blogDB.Close(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "; "))
}
