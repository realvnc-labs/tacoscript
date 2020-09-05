// +build !windows

package apptest

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func AssertFileMatchesExpectationOS(filePath string, fe *FileExpectation) (isMatched bool, reason string, err error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, "", err
	}

	if fe.ExpectedMode > 0 && fe.ExpectedMode != info.Mode() {
		return false, fmt.Sprintf("file '%s' has mode '%v' but '%v' was expected", filePath, info.Mode(), fe.ExpectedMode), nil
	}

	if fe.ExpectedGroup == "" && fe.ExpectedUser == "" {
		return true, "", nil
	}

	var fileUser string
	var fileGroup string
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		fileUserI, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
		if err != nil {
			return false, "", err
		}
		fileUser = fileUserI.Name

		fileGroupI, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
		if err != nil {
			return false, "", err
		}
		fileGroup = fileGroupI.Name
	}

	if fe.ExpectedUser != "" && fileUser != fe.ExpectedUser {
		return false, fmt.Sprintf("file '%s' should have '%s' as owner but has '%s'", filePath, fe.ExpectedUser, fileUser), nil
	}

	if fe.ExpectedGroup != "" && fileGroup != fe.ExpectedGroup {
		return false, fmt.Sprintf("file '%s' should have '%s' as group but has '%s'", filePath, fe.ExpectedGroup, fileGroup), nil
	}

	return true, "", nil
}
