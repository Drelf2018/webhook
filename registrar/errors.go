package registrar

import "errors"

var (
	ErrNoRegistrar = errors.New("webhook/registrar: registrar is nil")
	ErrNoAuth      = errors.New("webhook/registrar: no Authorization is provided")
	ErrNoUID       = errors.New("webhook/registrar: no uid is provided")
	ErrNoPassword  = errors.New("webhook/registrar: no password is provided")
	ErrVerify      = errors.New("webhook/registrar: verification failure")
	ErrUIDMismatch = errors.New("webhook/registrar: uid mismatch")
)
