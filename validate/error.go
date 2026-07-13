package validate

import "github.com/lasfh/goapi/logs/resperr"

type validateError struct {
	*resperr.ResponseError

	details map[string]string
}

func (v *validateError) Error() string {
	return v.ResponseError.Error()
}

func (v *validateError) Unwrap() []error {
	return v.ResponseError.Unwrap()
}

func (v *validateError) StatusCode() int {
	return v.ResponseError.StatusCode()
}

func (v *validateError) Message() string {
	return v.ResponseError.Message()
}

func (v *validateError) Details() map[string]string {
	return v.details
}
