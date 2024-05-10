package dao

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/gorms"
	"github.com/Drelf2018/webhook/database/model"
)

var postDB, userDB *gorms.DB

func Open(postPath, userPath string) error {
	err := os.MkdirAll(filepath.Dir(postPath), os.ModePerm)
	if err != nil {
		return err
	}

	postDB = gorms.SetSQLite(postPath).AutoMigrate(&model.Tag{}, &model.Post{})
	if postDB.Error() != nil {
		return postDB.Error()
	}

	err = os.MkdirAll(filepath.Dir(userPath), os.ModePerm)
	if err != nil {
		return err
	}

	userDB = gorms.SetSQLite(userPath).AutoMigrate(&model.Job{}, &model.User{})
	if userDB.Error() != nil {
		return userDB.Error()
	}

	return nil
}

func Close() error {
	err := userDB.Close()
	if err != nil {
		return err
	}
	return postDB.Close()
}
