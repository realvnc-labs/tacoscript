package conv_test

import (
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/conv"
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
			name:        "error when no units",
			inputStr:    "100",
			expectedErr: conv.ErrFileSizeInvalidUnits,
		},
		{
			name:        "error when unknown units",
			inputStr:    "100t",
			expectedErr: conv.ErrFileSizeInvalidUnits,
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
