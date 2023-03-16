//go:build !windows
// +build !windows

package realvnc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrMissingConfigValueName  = errors.New("missing config value name")
	ErrConfigWriterCannotBeNil = errors.New("config writer cannot be nil")
)

type ConfigValue struct {
	Name  string
	Value string
}

type ConfigValues struct {
	scanner      *bufio.Scanner
	configWriter io.Writer
}

func NewConfigValuesFromFile(configFilename string) (configValues *ConfigValues, err error) {
	configBytes, err := os.ReadFile(configFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if we didn't find the file then still return valid config values with the err.
			// we'll still be able to proceed with adding config values to a new file.
			return &ConfigValues{}, err
		}
		return nil, err
	}

	configReader := bytes.NewReader(configBytes)
	configValues, err = NewConfigValuesFromReader(configReader)
	if err != nil {
		return nil, err
	}

	return configValues, nil
}

func NewConfigValuesFromReader(reader io.Reader) (configValues *ConfigValues, err error) {
	scanner := bufio.NewScanner(reader)

	configValues = &ConfigValues{
		scanner: scanner,
	}

	return configValues, nil
}

func (cv *ConfigValues) HasScanner() bool {
	return cv.scanner != nil
}

func (cv *ConfigValues) Scan() bool {
	return cv.scanner.Scan()
}

func (cv *ConfigValues) Text() string {
	return cv.scanner.Text()
}

func (cv *ConfigValues) SetOutputWriter(writer io.Writer) {
	cv.configWriter = writer
}

func (cv *ConfigValues) GetOutputWriter() (writer io.Writer) {
	return cv.configWriter
}

func (cv *ConfigValues) WriteValue(configValue ConfigValue) (err error) {
	outputLine := fmt.Sprintf("%s=%s", configValue.Name, configValue.Value)
	return cv.WriteLine(outputLine)
}

func (cv *ConfigValues) WriteLine(outputLine string) (err error) {
	if cv.configWriter == nil {
		return ErrConfigWriterCannotBeNil
	}
	_, err = cv.configWriter.Write([]byte(outputLine + "\n"))
	return err
}

func ApplyConfigChange(existingValue ConfigValue, inputValue ConfigValue) (
	updated bool, updatedVal ConfigValue, err error) {
	if inputValue.Value == existingValue.Value {
		return false, existingValue, nil
	}
	return true, inputValue, nil
}

func ParseConfigKeyValueLine(inputLine string) (skip bool, configValue ConfigValue, err error) {
	parts := strings.Split(inputLine, "=")
	if len(parts) == 1 {
		return true, configValue, nil
	}

	trimmedName := strings.TrimSpace(parts[0])
	trimmedValue := strings.TrimSpace(parts[1])

	// return err if we don't have a key
	if trimmedName == "" {
		return true, configValue, ErrMissingConfigValueName
	}

	configValue = ConfigValue{
		Name:  trimmedName,
		Value: trimmedValue,
	}

	return false, configValue, nil
}
