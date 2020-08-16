package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type FsManager interface {
	FileExists(filePath string) (bool, error)
}

type FsManagerMock struct {
	CalledFilePaths []string
	ErrToReturn     error
	ExistsToReturn  bool
}

func (fmm *FsManagerMock) FileExists(filePath string) (bool, error) {
	fmm.CalledFilePaths = append(fmm.CalledFilePaths, filePath)
	return fmm.ExistsToReturn, fmm.ErrToReturn
}

type OSFsManager struct {}

func (fmm *OSFsManager) FileExists(filePath string) (bool, error) {
	if filePath == "" {
		return false, nil
	}

	logrus.Debugf("will check if file '%s' is missing", filePath)
	_, e := os.Stat(filePath)
	if e == nil {
		return true, nil
	}

	if os.IsNotExist(e) {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if file '%s' exists: %w", filePath, e)
}
