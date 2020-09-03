package apptest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ErrorExpectation struct {
	PartialText string
	FullText    string
	FullError   error
}

func AssertErrorExpectation(t *testing.T, err error, exp *ErrorExpectation) {
	if exp.FullError != nil {
		assert.Equal(t, err, exp.FullError)
		return
	}

	if exp.FullText != "" {
		assert.EqualError(t, err, exp.FullText)
		return
	}

	if exp.PartialText != "" {
		errorTextToCheck := ""
		if err != nil {
			errorTextToCheck = err.Error()
		}

		assert.Contains(t, errorTextToCheck, exp.PartialText)
		return
	}

	assert.NoError(t, err)
}
