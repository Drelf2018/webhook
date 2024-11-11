package registrar

import "errors"

var (
	ErrNoRegistrar = errors.New("webhook/api/registrar: registrar is nil")
	ErrNoAuth      = errors.New("webhook/api/registrar: no Authorization is provided")
	ErrNoUID       = errors.New("webhook/api/registrar: no uid is provided")
	ErrNoPassword  = errors.New("webhook/api/registrar: no password is provided")
	ErrVerify      = errors.New("webhook/api/registrar: verification failure")
	ErrUIDMismatch = errors.New("webhook/api/registrar: uid mismatch")
)
