//go:build windows
// +build windows

package realvncserver

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/winreg"
)

const (
	DefaultWindowsExecPath = `C:\Program Files\RealVNC\VNC Server`
	DefaultWindowsExecName = `vncserver.exe`
)

var (
	HKLMBaseKey = `HKLM:\SOFTWARE\RealVNC\vncserver`
	HKCUBaseKey = `HKCU:\Software\RealVNC\vncserver`
)

func (rvste *RealVNCServerTaskExecutor) applyConfigChanges(rvst *RealVNCServerTask) (addedCount int, updatedCount int, err error) {
	baseKey := getBaseKeyForServerMode(rvst.ServerMode)

	err = rvst.tracker.WithNewValues(func(fieldName string, fs FieldStatus) (err error) {
		regPath := fieldName
		regValue, err := rvst.getFieldValueAsString(fieldName)
		if err != nil {
			return err
		}

		desc := ""

		if fs.Clear {
			_, desc, err = winreg.RemoveValue(baseKey, regPath)
			if err != nil {
				return err
			}
			if strings.Contains(desc, "removed") {
				updatedCount++
				logrus.Debugf(`removed key %s\%s`, baseKey, regPath)
			}
		} else {
			_, desc, err = winreg.SetValue(baseKey, regPath, regValue, winreg.REG_SZ)
			if err != nil {
				return err
			}
			if strings.Contains(desc, "added") {
				addedCount++
				logrus.Debugf(`added key %s\%s with %s`, baseKey, regPath, regValue)
			} else if strings.Contains(desc, "updated") {
				updatedCount++
				logrus.Debugf(`updated key %s\%s with %s`, baseKey, regPath, regValue)
			}
		}

		err = rvst.tracker.SetChangeApplied(fieldName)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return addedCount, updatedCount, nil
}

func (rvste *RealVNCServerTaskExecutor) ReloadConfig(rvst *RealVNCServerTask) (err error) {
	var cmd *exec.Cmd

	cmdLine := rvste.makeReloadPSCmdLine(rvst)

	cmd = exec.Command("powershell", cmdLine)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	cmdRunner := tacoexec.OSApi{}
	err = cmdRunner.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed reloading vnc server configuration: %w", err)
	}

	return nil
}

func getBaseKeyForServerMode(serverMode string) (baseKey string) {
	baseKey = HKLMBaseKey
	if serverMode == UserServerMode {
		baseKey = HKCUBaseKey
	}
	return baseKey
}

func (rvste *RealVNCServerTaskExecutor) makeReloadPSCmdLine(rvst *RealVNCServerTask) (cmdLine string) {
	baseCmdLine := `Start-Process -FilePath '%s\%s' -WindowStyle Hidden  -ArgumentList '%s'`
	argumentList := `service -reload`

	if rvst.ServerMode == UserServerMode {
		argumentList = `-reload`
	}

	cmdLine = fmt.Sprintf(baseCmdLine, DefaultWindowsExecPath, DefaultWindowsExecName, argumentList)
	return cmdLine
}
