//go:build windows
// +build windows

package utils

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
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
	log.Debug("no chown support under windows")

	return nil
}
