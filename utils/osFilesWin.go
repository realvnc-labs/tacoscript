// +build windows

package utils

import (
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
