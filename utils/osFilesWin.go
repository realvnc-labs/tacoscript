//go:build windows
// +build windows

package utils

import (
	"fmt"
	"os"
	"strings"
)

func ParseLocationOS(rawLocation string) string {
	if !strings.HasPrefix(rawLocation, "file:") {
		return rawLocation
	}

	rawLocation = strings.TrimPrefix(rawLocation, "file:///")
	rawLocation = strings.Replace(rawLocation, "/", string(os.PathSeparator), -1)

	return rawLocation
}

func Chown(targetFilePath, userName, groupName string) error {
	return fmt.Errorf("no chown support under windows")
}
