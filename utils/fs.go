package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/secsy/goftp"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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

func MoveFile(sourceFilePath, targetFilePath string) error {
	err := os.Rename(sourceFilePath, targetFilePath)
	return err
}

func CopyLocalFile(sourceFilePath, targetFilePath string) error {
	input, err := ioutil.ReadFile(sourceFilePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(targetFilePath, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func DownloadHttpFile(url *url.URL, targetFilePath string) error {
	out, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer CloseResourceSecure(targetFilePath, out)

	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer CloseResourceSecure("http body", resp.Body)

	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadHttpsFile(skipTls bool, url *url.URL, targetFilePath string) error {
	out, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer CloseResourceSecure(targetFilePath, out)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTls,
			},
		},
	}

	resp, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer CloseResourceSecure("http body", resp.Body)

	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadFtpFile(url *url.URL, targetFilePath string) error {
	ftpCfg := goftp.Config{}
	if url.User != nil {
		usrlLogin := url.User.Username()
		usrPass, passIsSet := url.User.Password()
		if passIsSet {
			ftpCfg.User = usrlLogin
			ftpCfg.Password = usrPass
		}
	}

	ftpClient, err := goftp.DialConfig(ftpCfg, url.Host)

	if err != nil {
		return err
	}

	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}

	err = ftpClient.Retrieve(url.Path, targetFile)
	if err != nil {
		return err
	}

	return nil
}
