package errs

import "fmt"

type I18nError struct {
	Err  error
	Code string
}

func NewI18NError(format string, err error, code string) I18nError {
	err = fmt.Errorf(format, err)
	return I18nError{err, code}
}

func (e I18nError) Error() string {
	return e.Err.Error()
}

func (e I18nError) Unwrap() error {
	return e.Err
}
