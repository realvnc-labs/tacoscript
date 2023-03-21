package apptest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/realvnc-labs/tacoscript/utils"
)

func DeleteFiles(files []string) error {
	errs := &utils.Errors{
		Errs: []error{},
	}
	for _, file := range files {
		errs.Add(DeleteFileIfExists(file))
	}

	return errs.ToError()
}

func DeleteFileIfExists(filePath string) error {
	fileExists, err := utils.FileExists(filePath)
	if err != nil {
		return err
	}
	if !fileExists {
		return nil
	}

	return os.Remove(filePath)
}

type FileExpectation struct {
	ShouldExist      bool
	ExpectedMode     os.FileMode
	FilePath         string
	ExpectedContent  string
	ExpectedUser     string
	ExpectedGroup    string
	ExpectedEncoding string
}

func AssertFileMatchesExpectation(fe *FileExpectation) (isExpectationMatched bool, nonMatchedReason string, err error) {
	fileExists, err := utils.FileExists(fe.FilePath)
	if err != nil {
		return false, "", err
	}

	if fe.ShouldExist && !fileExists {
		return false, fmt.Sprintf("file '%s' doesn't exist but it should", fe.FilePath), nil
	}

	if !fe.ShouldExist && fileExists {
		return false, fmt.Sprintf("file '%s' exists but it shouldn't", fe.FilePath), nil
	}

	if !fe.ShouldExist && !fileExists {
		return true, "", nil
	}

	fileContentsBytes, err := os.ReadFile(fe.FilePath)
	if err != nil {
		return false, "", err
	}

	fileContents := ""
	if fe.ExpectedEncoding != "" {
		fileContents, err = utils.Decode(fe.ExpectedEncoding, fileContentsBytes)
		if err != nil {
			return false, "", err
		}
	} else {
		fileContents = string(fileContentsBytes)
	}

	if fe.ExpectedContent != fileContents {
		return false,
			fmt.Sprintf("file Contents '%s' at '%s' didn't match the expected one '%s'",
				fileContents,
				fe.FilePath,
				fe.ExpectedContent,
			), nil
	}

	return AssertFileMatchesExpectationOS(fe.FilePath, fe)
}

type ChownInput struct {
	TargetFilePath string
	UserName       string
	GroupName      string
}

type FsManagerMock struct {
	FileExistsInputPath      []string
	FileExistsErrToReturn    error
	FileExistsExistsToReturn bool
	ChownInputs              []ChownInput
	ChownErrorToReturn       error
	StatInputName            []string
	StatOutputFileInfo       os.FileInfo
	StatOutputError          error
}

func (fmm *FsManagerMock) FileExists(filePath string) (bool, error) {
	fmm.FileExistsInputPath = append(fmm.FileExistsInputPath, filePath)
	return fmm.FileExistsExistsToReturn, fmm.FileExistsErrToReturn
}

func (fmm *FsManagerMock) Remove(filePath string) error {
	return nil
}

func (fmm *FsManagerMock) DownloadFile(ctx context.Context, targetLocation string, sourceURL *url.URL, skipTLSCheck bool) error {
	return nil
}

func (fmm *FsManagerMock) MoveFile(sourceFilePath, targetFilePath string) error {
	return nil
}

func (fmm *FsManagerMock) CopyLocalFile(sourceFilePath, targetFilePath string, mode os.FileMode) error {
	return nil
}

func (fmm *FsManagerMock) WriteFile(name, contents string, mode os.FileMode) error {
	return nil
}

func (fmm *FsManagerMock) ReadFile(filePath string) (content string, err error) {
	return "", nil
}

func (fmm *FsManagerMock) CreateDirPathIfNeeded(targetFilePath string, mode os.FileMode) error {
	return nil
}

func (fmm *FsManagerMock) Chmod(targetFilePath string, mode os.FileMode) error {
	return nil
}

func (fmm *FsManagerMock) ReadEncodedFile(encodingName, fileName string) (contentsUtf8 string, err error) {
	return
}

func (fmm *FsManagerMock) Chown(targetFilePath, userName, groupName string) error {
	fmm.ChownInputs = append(fmm.ChownInputs, ChownInput{
		TargetFilePath: targetFilePath,
		UserName:       userName,
		GroupName:      groupName,
	})

	return fmm.ChownErrorToReturn
}

func (fmm *FsManagerMock) Stat(name string) (os.FileInfo, error) {
	fmm.StatInputName = append(fmm.StatInputName, name)
	return fmm.StatOutputFileInfo, fmm.StatOutputError
}

// FakeFile implements FileLike and also os.FileInfo.
type FakeFile struct {
	Nam      string
	Contents string
	FileMode os.FileMode
}

func (f *FakeFile) Name() string {
	return f.Nam
}

func (f *FakeFile) Size() int64 {
	return int64(len(f.Contents))
}

func (f *FakeFile) Mode() os.FileMode {
	return f.FileMode
}

func (f *FakeFile) ModTime() time.Time {
	return time.Time{}
}

func (f *FakeFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *FakeFile) IsDir() bool {
	return false
}

func (f *FakeFile) Sys() interface{} {
	return nil
}
