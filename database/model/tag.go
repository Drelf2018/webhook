package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Key     string
	Value   string
	Comment string
	PostID  uint64
}
