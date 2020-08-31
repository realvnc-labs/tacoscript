package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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

type FileExpectation struct {
	FilePath         string
	ShouldExist      bool
	ExpectedContent  string
	ExpectedUser     string
	ExpectedGroup    string
	ExpectedMode     os.FileMode
	ExpectedEncoding string
}

type OSFsManager struct{}

func (fmm *OSFsManager) FileExists(filePath string) (bool, error) {
	return FileExists(filePath)
}

func FileExists(filePath string) (bool, error) {
	if filePath == "" {
		return false, nil
	}

	logrus.Debugf("will check if file '%s' exists", filePath)
	_, e := os.Stat(filePath)
	if e == nil {
		return true, nil
	}

	if os.IsNotExist(e) {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if file '%s' exists: %w", filePath, e)
}

func AssertFileMatchesExpectation(filePath string, fe *FileExpectation) (bool, string, error) {
	fileExists, err := FileExists(filePath)
	if err != nil {
		return false, "", err
	}

	if fe.ShouldExist && !fileExists {
		return false, fmt.Sprintf("file '%s' doesn't exist but it should", filePath), nil
	}

	if !fe.ShouldExist && fileExists {
		return false, fmt.Sprintf("file '%s' exists but it shouldn't", filePath), nil
	}

	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, "", err
	}

	fileContents := ""
	if fe.ExpectedEncoding != "" {
		fileContents, err = Decode(fe.ExpectedEncoding, fileContentsBytes)
		if err != nil {
			return false, "", err
		}
	} else {
		fileContents = string(fileContentsBytes)
	}

	if fe.ExpectedContent != fileContents {
		return false, fmt.Sprintf("file contents '%s' at '%s' didn't match the expected one '%s'", fileContents, filePath, fe.ExpectedContent), nil
	}

	return AssertFileMatchesExpectationOS(filePath, fe)
}
