package api

import "errors"

var (
	ErrAuthNotExist   = errors.New("webhook/api: Authorization does not exist")
	ErrBlogNotExist   = errors.New("webhook/api: blog does not exist")
	ErrUserNotExist   = errors.New("webhook/api: user does not exist")
	ErrTaskNotExist   = errors.New("webhook/api: task does not exist")
	ErrFilterNotExist = errors.New("webhook/api: filter does not exist")

	ErrExpired    = errors.New("webhook/api: token is expired")
	ErrPermDenied = errors.New("webhook/api: permission denied")

	ErrUserRegistered = errors.New("webhook/api: user registered")
	ErrIncorrectPwd   = errors.New("webhook/api: incorrect password")
	ErrBanned         = errors.New("webhook/api: user has been banned")
)
