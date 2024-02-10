package validator

import (
	"fmt"
	"regexp"
)

var (
	EmailRX = regexp.MustCompile("[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?")
)

type Validator struct {
	FieldErrors map[string]error
}

func New() *Validator {
	return &Validator{
		FieldErrors: make(map[string]error),
	}
}

func (v *Validator) HasErrors() bool {
	return len(v.FieldErrors) != 0
}

func (v *Validator) GetError(field string) error {
	return v.FieldErrors[field]
}

func (v *Validator) Error() string {
	msg := "validator failed for "
	for field := range v.FieldErrors {
		msg += field + ", "
	}
	return msg
}

func (v *Validator) AddError(field string, err error) {
	v.FieldErrors[field] = err
}

func (v *Validator) Check(predicate bool, field string, errMsg string) {
	if predicate {
		v.AddError(field, fmt.Errorf(errMsg))
	}
}

func (v *Validator) MatchesRX(rx *regexp.Regexp, value string, field string, errMsg string) {
	v.Check(!rx.MatchString(value), field, errMsg)
}
