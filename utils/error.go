package utils

import "strings"

type JoinError []error

func (e JoinError) Error() string {
	errs := make([]string, 0, len(e))
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "; ")
}

func (e JoinError) Unwrap() []error {
	return e
}
