package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequired(t *testing.T) {
	testCases := []struct {
		InputVal       string
		InputPath      string
		ExpectedErrMsg string
	}{
		{
			InputVal:       "123",
			ExpectedErrMsg: "",
		},
		{
			InputVal:       "",
			InputPath:      "somepath",
			ExpectedErrMsg: "empty required value at path 'somepath'",
		},
		{
			InputVal:       "   ",
			InputPath:      "somepath1",
			ExpectedErrMsg: "empty required value at path 'somepath1'",
		},
	}

	for _, testCase := range testCases {
		actualErr := ValidateRequired(testCase.InputVal, testCase.InputPath)
		if testCase.ExpectedErrMsg == "" {
			assert.NoError(t, actualErr)
			continue
		}

		assert.EqualError(t, actualErr, testCase.ExpectedErrMsg)
	}
}
