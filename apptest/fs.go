package apptest

import (
	"os"

	"github.com/cloudradar-monitoring/tacoscript/utils"
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
