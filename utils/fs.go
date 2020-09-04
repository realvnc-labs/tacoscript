package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/secsy/goftp"
	"github.com/sirupsen/logrus"
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
	ShouldExist      bool
	ExpectedMode     os.FileMode
	FilePath         string
	ExpectedContent  string
	ExpectedUser     string
	ExpectedGroup    string
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

func AssertFileMatchesExpectation(filePath string, fe *FileExpectation) (isExpectationMatched bool, nonMatchedReason string, err error) {
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

	if !fe.ShouldExist && !fileExists {
		return true, "", nil
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
		return false,
			fmt.Sprintf("file contents '%s' at '%s' didn't match the expected one '%s'",
				fileContents,
				filePath,
				fe.ExpectedContent,
			), nil
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

	err = ioutil.WriteFile(targetFilePath, input, 0600)
	if err != nil {
		return err
	}

	return nil
}

func DownloadHTTPFile(ctx context.Context, u fmt.Stringer, targetFilePath string) error {
	out, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer CloseResourceSecure(targetFilePath, out)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer CloseResourceSecure("http body", resp.Body)

	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadHTTPSFile(ctx context.Context, skipTLS bool, u fmt.Stringer, targetFilePath string) error {
	out, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer CloseResourceSecure(targetFilePath, out)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLS,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer CloseResourceSecure("http body", resp.Body)

	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadFtpFile(ctx context.Context, u *url.URL, targetFilePath string) error {
	ftpCfg := goftp.Config{}
	if u.User != nil {
		usrlLogin := u.User.Username()
		usrPass, passIsSet := u.User.Password()
		if passIsSet {
			ftpCfg.User = usrlLogin
			ftpCfg.Password = usrPass
		}
	}

	ftpClient, err := goftp.DialConfig(ftpCfg, u.Host)

	if err != nil {
		return err
	}

	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}

	err = ftpClient.Retrieve(u.Path, targetFile)
	if err != nil {
		return err
	}

	return nil
}

func CreateDirPathIfNeeded(targetFilePath string, mode os.FileMode) error {
	targetFileDir := filepath.Dir(targetFilePath)
	if targetFileDir != "" {
		logrus.Debugf("will create dirs tree '%s", targetFileDir)
		return os.MkdirAll(targetFileDir, mode)
	}

	return nil
}

func DownloadFile(ctx context.Context, targetLocation string, sourceURL *url.URL, skipTLSCheck bool) error {
	logrus.Debug("source location is a remote file path")

	var err error
	switch sourceURL.Scheme {
	case "http":
		err = DownloadHTTPFile(ctx, sourceURL, targetLocation)
	case "https":
		err = DownloadHTTPSFile(ctx, skipTLSCheck, sourceURL, targetLocation)
	case "ftp":
		err = DownloadFtpFile(ctx, sourceURL, targetLocation)
	default:
		err = fmt.Errorf(
			"unknown or unsupported protocol '%s' to download data from '%s'",
			sourceURL.Scheme,
			sourceURL,
		)
	}

	if err != nil {
		return err
	}

	return nil
}
