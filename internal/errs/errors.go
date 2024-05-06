package errs

import (
	"errors"
)

var (
	ErrinternalError = errors.New("unexpected internal error")
	ErrInvalidInput  = errors.New("invalidinput")
	ErrNotFound      = errors.New("resourcenotfound")
	ErrNotAuthorized = errors.New("unauthorized")
	ErrConflict      = errors.New("there is a conflict with the current status")
)
