package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Key    string
	Value  string
	Hint   string
	PostID uint64
}
