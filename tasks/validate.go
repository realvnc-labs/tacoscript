package tasks

import (
	"fmt"
	"strings"
)

func ValidateRequired(val, path string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("empty required value at path '%s'", path)
	}

	return nil
}

func ValidateRequiredMany(vals []string, path string) error {
	for _, val := range vals {
		if strings.TrimSpace(val) != "" {
			return nil
		}
	}

	return fmt.Errorf("empty required values at path '%s'", path)
}
