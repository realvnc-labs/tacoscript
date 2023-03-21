package conv_test

import (
	"testing"

	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/stretchr/testify/assert"
)

func TestShouldConvertFileSizeToUInt64(t *testing.T) {
	cases := []struct {
		name          string
		inputStr      string
		expectedValue uint64
		expectedErr   error
	}{
		{
			name:        "error when unknown units",
			inputStr:    "100t",
			expectedErr: conv.ErrFileSizeInvalidUnits,
		},
		{
			name:          "assume bytes",
			inputStr:      "100",
			expectedValue: 100,
		},
		{
			name:          "100b",
			inputStr:      "100b",
			expectedValue: 100,
		},
		{
			name:          "100k",
			inputStr:      "100k",
			expectedValue: uint64(100 * 1024),
		},
		{
			name:          "200m",
			inputStr:      "200m",
			expectedValue: uint64(200 * 1024 * 1024),
		},
		{
			name:          "300g",
			inputStr:      "300g",
			expectedValue: uint64(300 * 1024 * 1024 * 1024),
		},
		{
			name:        "error when not a number, but valid units",
			inputStr:    "bbb",
			expectedErr: conv.ErrNotANumber,
		},
		{
			name:        "error when not a number",
			inputStr:    "bbc",
			expectedErr: conv.ErrNotANumber,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := conv.ConvertToFileSize(tc.inputStr)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.Equal(t, tc.expectedValue, val)
			}
		})
	}
}
