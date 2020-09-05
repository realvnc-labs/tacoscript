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

type FsManager struct{}

func (fmm *FsManager) FileExists(filePath string) (bool, error) {
	return FileExists(filePath)
}

func (fmm *FsManager) Remove(filePath string) error {
	return os.Remove(filePath)
}

func (fmm *FsManager) DownloadFile(ctx context.Context, targetLocation string, sourceURL *url.URL, skipTLSCheck bool) error {
	return DownloadFile(ctx, targetLocation, sourceURL, skipTLSCheck)
}

func (fmm *FsManager) MoveFile(sourceFilePath, targetFilePath string) error {
	return MoveFile(sourceFilePath, targetFilePath)
}

func (fmm *FsManager) CopyLocalFile(sourceFilePath, targetFilePath string) error {
	return CopyLocalFile(sourceFilePath, targetFilePath)
}

func (fmm *FsManager) WriteFile(name, contents string, mode os.FileMode) error {
	return ioutil.WriteFile(name, []byte(contents), mode)
}

func (fmm *FsManager) ReadFile(filePath string) (content string, err error) {
	contentsByte, err := ioutil.ReadFile(filePath)

	return string(contentsByte), err
}

func (fmm *FsManager) CreateDirPathIfNeeded(targetFilePath string, mode os.FileMode) error {
	return CreateDirPathIfNeeded(targetFilePath, mode)
}

func (fmm *FsManager) Chmod(targetFilePath string, mode os.FileMode) error {
	return os.Chmod(targetFilePath, mode)
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
