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
