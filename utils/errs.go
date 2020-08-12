package utils

import (
	"errors"
	"strings"
)

type Errors struct {
	Errs []error
}

func (ve *Errors) Add(err error) {
	if err == nil {
		return
	}

	ve.Errs = append(ve.Errs, err)
}

func (ve Errors) ToError() error {
	if len(ve.Errs) == 0 {
		return nil
	}

	rawErrors := make([]string, 0, len(ve.Errs))
	for _, err := range ve.Errs {
		rawErrors = append(rawErrors, err.Error())
	}

	return errors.New(strings.Join(rawErrors, ", "))
}
