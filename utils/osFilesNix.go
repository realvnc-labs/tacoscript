//go:build !windows
// +build !windows

package utils

import (
	"os"
	"os/user"
	"strconv"
	"strings"
)

func ParseLocationOS(rawLocation string) string {
	if !strings.HasPrefix(rawLocation, "file:") {
		return rawLocation
	}

	rawLocation = strings.TrimPrefix(rawLocation, "file://")

	return rawLocation
}

func Chown(targetFilePath, userName, groupName string) error {
	usrID, groupID := -1, -1

	if userName != "" {
		sysUser, err := user.Lookup(userName)
		if err != nil {
			return err
		}
		usrID, err = strconv.Atoi(sysUser.Uid)
		if err != nil {
			return err
		}
	}

	if groupName != "" {
		sysGroup, err := user.LookupGroup(groupName)
		if err != nil {
			return err
		}

		groupID, err = strconv.Atoi(sysGroup.Gid)
		if err != nil {
			return err
		}
	}

	return os.Chown(targetFilePath, usrID, groupID)
}
