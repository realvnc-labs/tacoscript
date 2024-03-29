//go:build !windows
// +build !windows

package realvncserver

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/tasks/support/realvnc"
	"github.com/realvnc-labs/tacoscript/utils"
)

const (
	DefaultMacConfigReloadCmd    = `/Library/vnc/vncserver`
	DefaultLinuxConfigReloadCmd  = `/usr/bin/vncserver-x11`
	DefaultLinuxLicenseReloadCmd = `/usr/bin/vnclicense`

	DefaultConfigFilePermissions = 0644
)

func (rvste *Executor) applyConfigChanges(rvst *Task) (addedCount int, updatedCount int, err error) {
	configValues, outputBuffer, err := newConfigValuesWithOutputBuffer(rvst)
	if err != nil {
		return 0, 0, err
	}

	addedCount, updatedCount, err = rvste.makeChanges(rvst, configValues)
	if err != nil {
		return 0, 0, err
	}

	if addedCount > 0 || updatedCount > 0 {
		err = commitChanges(rvst, outputBuffer)
		if err != nil {
			return 0, 0, err
		}
	}

	return addedCount, updatedCount, nil
}

func newConfigValuesWithOutputBuffer(rvst *Task) (
	configValuesFile *realvnc.ConfigValues, outputBuffer *bytes.Buffer, err error) {
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

	outputBuffer = &bytes.Buffer{}
	configValuesFile.SetOutputWriter(outputBuffer)

	return configValuesFile, outputBuffer, nil
}

func (rvste *Executor) makeChanges(rvst *Task, configValues *realvnc.ConfigValues) (
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

func (rvste *Executor) updateExistingValues(rvst *Task, configValues *realvnc.ConfigValues) (
	updatedCount int, err error) {
	updatedCount = 0
	lineNum := 0

	// if no scanner then we aren't reading an existing config file so no values to update
	if !configValues.HasScanner() {
		return 0, nil
	}

	logrus.Debugf("checking for config values to update")

	for configValues.Scan() {
		updated := false
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

		fieldStatus, found := rvst.fieldTracker.GetFieldStatus(fieldName)
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

			if existingConfigValue.Value != changeValue.Value {
				updated = true
				logrus.Debugf(`updated %s with %s`, changeValue.Name, changeValue.Value)
			}
		} else {
			updated = true
			logrus.Debugf(`removed %s`, fieldName)
		}

		err = rvst.fieldTracker.SetChangeApplied(fieldName)
		if err != nil {
			return 0, fmt.Errorf("failed to update change status %s: %v", fieldName, err)
		}

		if updated {
			updatedCount++
		}
	}

	return updatedCount, nil
}

func (rvste *Executor) addNewValues(rvst *Task, configValues *realvnc.ConfigValues) (
	addedCount int, err error) {
	addedCount = 0

	logrus.Debugf("checking for new config values")

	err = rvst.fieldTracker.WithNewValues(func(fieldName string, fs fieldstatus.FieldStatus) (err error) {
		// ignore fields where the change has been applied already (aka updating/cleared fields)
		if fs.ChangeApplied || fs.Clear {
			return nil
		}

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
			return fmt.Errorf("failed to added config value %s: %v", fieldName, err)
		}

		logrus.Debugf(`added %s with %s`, newValue.Name, newValue.Value)

		addedCount++
		return nil
	})

	if err != nil {
		return 0, err
	}
	return addedCount, nil
}

func (t *Task) getChangeValue(fieldName string) (changeValue realvnc.ConfigValue, err error) {
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

func commitChanges(rvst *Task, outputBuffer *bytes.Buffer) (err error) {
	configFilename := rvst.ConfigFile
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

	perms := fs.FileMode(DefaultConfigFilePermissions)
	if existingConfig {
		// preserve existing config file permissions
		perms = info.Mode().Perm()
	}

	err = os.WriteFile(configFilename, outputBuffer.Bytes(), perms)
	if err != nil {
		return err
	}

	// TODO: (rs): what about file owners? need to consider Windows too.
	err = os.Chmod(configFilename, perms)
	if err != nil {
		return err
	}

	logrus.Debugf("wrote config file at %s", configFilename)
	return nil
}

func (rvste *Executor) ReloadConfig(rvst *Task) (err error) {
	var cmd *exec.Cmd

	execCmd, params := MakeReloadCmdLine(rvst, runtime.GOOS)

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

	logrus.Debugf("config reloaded successfully")

	return nil
}

func MakeReloadCmdLine(rvst *Task, goos string) (cmd string, params []string) {
	argumentList := []string{`-service`, `-reload`}

	if rvst.ReloadExecPath == "" {
		if goos == "darwin" {
			cmd = DefaultMacConfigReloadCmd
		} else {
			// linux only
			if rvst.UseVNCLicenseReload {
				cmd = DefaultLinuxLicenseReloadCmd
			} else {
				cmd = DefaultLinuxConfigReloadCmd
			}
		}
	} else {
		cmd = rvst.ReloadExecPath
		logrus.Debugf(`user specified reload_exec_path = %s`, cmd)
	}

	if rvst.ServerMode == UserServerMode ||
		goos != "darwin" && rvst.UseVNCLicenseReload {
		argumentList = []string{`-reload`}
	}

	return cmd, argumentList
}
