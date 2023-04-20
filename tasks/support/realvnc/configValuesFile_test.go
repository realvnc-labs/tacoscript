//go:build !windows
// +build !windows

package realvnc

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSkipNonKeyValueLines(t *testing.T) {
	cases := []struct {
		name      string
		inputLine string
	}{
		{
			name:      "comment line",
			inputLine: "  # this is a comment",
		},
		{
			name:      "rogue line",
			inputLine: "this is a rogue line",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			skipLine, _, _ := ParseConfigKeyValueLine(tc.inputLine)
			require.True(t, skipLine)
		})
	}
}

func TestShouldReadValidKeyValueLines(t *testing.T) {
	cases := []struct {
		name          string
		inputLine     string
		expectedName  string
		expectedValue string
		expectedErr   error
	}{
		{
			name:          "basic valid line",
			inputLine:     "ValueName=field_value",
			expectedName:  "ValueName",
			expectedValue: "field_value",
		},
		{
			name:        "invalid line, no name",
			inputLine:   "=field_value",
			expectedErr: ErrMissingConfigValueName,
		},
		{
			name:          "mixed case name, unchanged",
			inputLine:     "Value_name=field_value",
			expectedName:  "Value_name",
			expectedValue: "field_value",
		},
		{
			name:          "mixed case value, unchanged",
			inputLine:     "value_name=FIEld_value",
			expectedName:  "value_name",
			expectedValue: "FIEld_value",
		},
		{
			name:          "white space trimmed around name",
			inputLine:     "   \tvalue_name   =value",
			expectedName:  "value_name",
			expectedValue: "value",
		},
		{
			name:          "white space trimmed around value",
			inputLine:     "value_name=    \t   value   \t",
			expectedName:  "value_name",
			expectedValue: "value",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			skipLine, configValue, err := ParseConfigKeyValueLine(tc.inputLine)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.False(t, skipLine)
				require.Equal(t, tc.expectedName, configValue.Name)
				require.Equal(t, tc.expectedValue, configValue.Value)
			} else {
				require.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}

func TestShouldUpdateModifiedConfigValue(t *testing.T) {
	inputConfigValue := ConfigValue{Name: "123", Value: "789"}
	existingConfigValue := ConfigValue{Name: "123", Value: "456"}

	updated, updatedVal, err := ApplyConfigChange(existingConfigValue, inputConfigValue)
	require.NoError(t, err)
	require.Equal(t, inputConfigValue, updatedVal)
	require.True(t, updated)
}

func TestShouldNotUpdateUnchangedConfigValue(t *testing.T) {
	inputConfigValue := ConfigValue{Name: "123", Value: "456"}
	existingConfigValue := ConfigValue{Name: "123", Value: "456"}

	updated, updatedVal, err := ApplyConfigChange(existingConfigValue, inputConfigValue)
	require.NoError(t, err)
	require.Equal(t, existingConfigValue, updatedVal)
	require.False(t, updated)
}

func TestShouldGetConfigValues(t *testing.T) {
	configFilename := "../../../testdata/realvncserver-config.conf.orig"

	configValues, err := NewConfigValuesFromFile(configFilename)
	require.NoError(t, err)
	require.NotNil(t, configValues)

	count := 0
	for configValues.Scan() {
		line := configValues.Text()
		assert.NotEmpty(t, line)
		count++
	}

	assert.Equal(t, 6, count)
}

func TestShouldErrorWhenCannotOpenConfigFile(t *testing.T) {
	configFilename := "../../../testdata/realvncserver-config-1.conf"

	_, err := NewConfigValuesFromFile(configFilename)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestShouldWriteConfigValueToWriter(t *testing.T) {
	configValues := &ConfigValues{}

	buf := new(bytes.Buffer)
	configValues.SetOutputWriter(buf)

	err := configValues.WriteLine("testline")
	require.NoError(t, err)
	err = configValues.WriteValue(ConfigValue{Name: "Test", Value: "Value"})
	require.NoError(t, err)

	results := buf.String()

	assert.Contains(t, results, "testline")
	assert.Contains(t, results, "Test=Value")
}
