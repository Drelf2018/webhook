package api

import (
	"github.com/Drelf2018/webhook/utils"
	"gorm.io/gorm"
)

var UserDB, BlogDB *gorm.DB

func CloseDB() error {
	errs := make(utils.JoinError, 0, 2)

	if userDB, err := UserDB.DB(); err != nil {
		errs = append(errs, err)
	} else if err = userDB.Close(); err != nil {
		errs = append(errs, err)
	}

	if blogDB, err := BlogDB.DB(); err != nil {
		errs = append(errs, err)
	} else if err = blogDB.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
