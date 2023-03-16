//go:build !windows
// +build !windows

package tasks

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/realvnc"
	"github.com/cloudradar-monitoring/tacoscript/utils"
)

const (
	DefaultMacExecPath   = `/Library/vnc`
	DefaultMacExecName   = `vncserver`
	DefaultLinuxExecPath = `/usr/bin`
	DefaultLinuxExecCmd  = `vncserver-x11`

	DefaultConfigFilePermissions = 0644
)

func (rvste *RealVNCServerTaskExecutor) applyConfigChanges(rvst *RealVNCServerTask) (addedCount int, updatedCount int, err error) {
	configValues, tempFile, err := newConfigValuesWithTempOutputFile(rvst)
	if err != nil {
		return 0, 0, err
	}

	// make sure to close the writer
	defer func() {
		closeErr := tempFile.Close()
		// if we're already returning an err then that's more important than the closeErr
		if err == nil {
			// if no existing err and a closeErr then return the closeErr
			err = closeErr
		}
	}()

	addedCount, updatedCount, err = rvste.makeChanges(rvst, configValues)
	if err != nil {
		return 0, 0, err
	}

	err = commitChanges(rvst, tempFile)
	if err != nil {
		return 0, 0, err
	}

	return addedCount, updatedCount, nil
}

func newConfigValuesWithTempOutputFile(rvst *RealVNCServerTask) (
	configValuesFile *realvnc.ConfigValues, tempFile *os.File, err error) {
	configFilename := rvst.ConfigFile
	if configFilename == "" {
		return nil, nil, errors.New(ErrConfigFileMustBeSpecifiedMsg)
	}

	logrus.Debugf("reading config values from %s", configFilename)
	configValuesFile, err = realvnc.NewConfigValuesFromFile(configFilename)
	if err != nil {
		// we can continue ok if no existing config file
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
	}

	baseConfigFilename := filepath.Base(rvst.ConfigFile)
	tempFile, err = os.CreateTemp("", baseConfigFilename)
	if err != nil {
		return nil, tempFile, err
	}
	logrus.Debugf("created temp config file at %s", tempFile.Name())

	configValuesFile.SetOutputWriter(tempFile)

	return configValuesFile, tempFile, nil
}

func (rvste *RealVNCServerTaskExecutor) makeChanges(rvst *RealVNCServerTask, configValues *realvnc.ConfigValues) (
	addedCount int, updatedCount int, err error) {
	updatedCount, err = rvste.updateExistingValues(rvst, configValues)
	if err != nil {
		return 0, 0, err
	}

	addedCount, err = rvste.addNewValues(rvst, configValues)
	if err != nil {
		return 0, 0, err
	}

	return addedCount, updatedCount, nil
}

func (rvste *RealVNCServerTaskExecutor) updateExistingValues(rvst *RealVNCServerTask, configValues *realvnc.ConfigValues) (
	updatedCount int, err error) {
	updatedCount = 0
	lineNum := 0

	// if no scanner then we aren't reading an existing config file so no values to update
	if !configValues.HasScanner() {
		return 0, nil
	}

	logrus.Debugf("checking for config values to update")

	for configValues.Scan() {
		inputLine := configValues.Text()
		lineNum++

		skipLine, existingConfigValue, err := realvnc.ParseConfigKeyValueLine(inputLine)
		if err != nil {
			return 0, fmt.Errorf("failed to parse config file line %d: %v", lineNum, err)
		}

		// if the line doesn't match as a key value pair then just write the line untouched
		if skipLine {
			err = configValues.WriteLine(inputLine)
			if err != nil {
				return 0, fmt.Errorf("failed to write config file line %d: %v", lineNum, err)
			}
			continue
		}

		fieldName := existingConfigValue.Name

		fieldStatus, fieldKey, found := rvst.tracker.GetFieldStatusByName(fieldName)
		if err != nil {
			return 0, fmt.Errorf("error while finding field %s: %v", fieldName, err)
		}

		if !found {
			// if the key value pair isn't found then just write the line untouched
			err = configValues.WriteLine(inputLine)
			if err != nil {
				return 0, fmt.Errorf("failed to write config file value %s at line %d: %v", fieldName, lineNum, err)
			}
			continue
		}

		if !fieldStatus.HasNewValue {
			// if the key value pair isn't being updated then just write the line untouched
			err = configValues.WriteLine(inputLine)
			if err != nil {
				return 0, fmt.Errorf("failed to write config file value %s at line %d: %v", fieldName, lineNum, err)
			}
			continue
		}

		// if we're removing then we're not write any value so we just fall through and update the
		// change status and updatedCount

		if !fieldStatus.Clear {
			changeValue, err := rvst.getChangeValue(fieldName)
			if err != nil {
				return 0, err
			}

			// write the new value
			err = configValues.WriteValue(changeValue)
			if err != nil {
				return 0, fmt.Errorf("failed to write config value %s at line %d: %v", fieldName, lineNum, err)
			}
		}

		err = rvst.tracker.SetChangeApplied(fieldKey)
		if err != nil {
			return 0, fmt.Errorf("failed to update change status %s: %v", fieldName, err)
		}

		updatedCount++
	}

	return updatedCount, nil
}

func (rvste *RealVNCServerTaskExecutor) addNewValues(rvst *RealVNCServerTask, configValues *realvnc.ConfigValues) (
	addedCount int, err error) {
	addedCount = 0

	logrus.Debugf("checking for new config values")

	err = rvst.tracker.WithNewValues(func(fk string, fs FieldStatus) (err error) {
		// ignore fields where the change has been applied already (aka updating/cleared fields)
		if fs.ChangeApplied {
			return
		}

		fieldName := fs.Name

		// get the current field value as a string
		val, err := rvst.getFieldValueAsString(fieldName)
		if err != nil {
			return err
		}

		// make a new config value to be added
		newValue := realvnc.ConfigValue{
			Name:  fieldName,
			Value: val,
		}

		err = configValues.WriteValue(newValue)
		if err != nil {
			return fmt.Errorf("failed to added config value %s: %v", fk, err)
		}

		addedCount++
		return nil
	})

	if err != nil {
		return 0, err
	}
	return addedCount, nil
}

func (t *RealVNCServerTask) getChangeValue(fieldName string) (changeValue realvnc.ConfigValue, err error) {
	// get the current field value as a string
	val, err := t.getFieldValueAsString(fieldName)
	if err != nil {
		return changeValue, err
	}

	// make the new config value for the change
	changeValue = realvnc.ConfigValue{
		Name:  fieldName,
		Value: val,
	}

	return changeValue, nil
}

func commitChanges(rvst *RealVNCServerTask, tempFile *os.File) (err error) {
	configFilename := rvst.ConfigFile
	tempFilename := tempFile.Name()
	existingConfig := true

	info, err := os.Stat(configFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			existingConfig = false
		} else {
			return err
		}
	}

	if existingConfig {
		if rvst.SkipBackup {
			err = os.Remove(configFilename)
			if err != nil {
				return err
			}
		} else {
			backupFilename := utils.GetBackupFilename(configFilename, rvst.Backup)

			err = os.Rename(configFilename, backupFilename)
			if err != nil {
				return err
			}

			logrus.Debugf("wrote backup config file at %s", backupFilename)
		}
	}

	err = os.Rename(tempFilename, configFilename)
	if err != nil {
		return err
	}

	perms := fs.FileMode(DefaultConfigFilePermissions)
	if existingConfig {
		// preserve existing config file permissions
		perms = info.Mode().Perm()
	}

	err = os.Chmod(configFilename, perms)
	if err != nil {
		return err
	}

	logrus.Debugf("wrote config file at %s", configFilename)
	return nil
}

func (rvste *RealVNCServerTaskExecutor) ReloadConfig(rvst *RealVNCServerTask) (err error) {
	var cmd *exec.Cmd

	execCmd, params := rvste.makeReloadCmdLine(rvst)

	cmd = exec.Command(execCmd, params...)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	cmdRunner := tacoexec.OSApi{}
	err = cmdRunner.Run(cmd)
	if err != nil {
		logrus.Debugf("error during settings reload: %s", err)
		logrus.Debugf("stderr = %s", errBuf.String())
		return fmt.Errorf("failed reloading vnc server configuration: %w", err)
	}

	return nil
}

func (rvste *RealVNCServerTaskExecutor) makeReloadCmdLine(rvst *RealVNCServerTask) (cmd string, params []string) {
	baseCmdLine := `%s/%s`
	argumentList := []string{`-service`, `-reload`}
	if rvst.ServerMode == UserServerMode {
		argumentList = []string{`-reload`}
	}

	if runtime.GOOS == "darwin" {
		cmd = fmt.Sprintf(baseCmdLine, DefaultMacExecPath, DefaultMacExecName)
	} else {
		cmd = fmt.Sprintf(baseCmdLine, DefaultLinuxExecPath, DefaultLinuxExecCmd)
	}

	return cmd, argumentList
}
