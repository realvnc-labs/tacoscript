// +build !windows

package utils

import (
	"strings"
)

func ParseLocationOS(rawLocation string) string {
	if !strings.HasPrefix(rawLocation, "file:") {
		return rawLocation
	}

	rawLocation = strings.TrimPrefix(rawLocation, "file://")

	return rawLocation
}
