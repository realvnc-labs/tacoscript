package tasks

import (
	"context"
	"net/url"
	"os"
)

type FsManager interface {
	FileExists(filePath string) (bool, error)
	Remove(filePath string) error
	DownloadFile(ctx context.Context, targetLocation string, sourceURL *url.URL, skipTLSCheck bool) error
	MoveFile(sourceFilePath, targetFilePath string) error
	CopyLocalFile(sourceFilePath, targetFilePath string) error
	WriteFile(name, contents string, mode os.FileMode) error
	ReadFile(filePath string) (content string, err error)
	CreateDirPathIfNeeded(targetFilePath string, mode os.FileMode) error
	Chmod(targetFilePath string, mode os.FileMode) error
	Chown(targetFilePath string, userName, groupName string) error
	Stat(name string) (os.FileInfo, error)
}
