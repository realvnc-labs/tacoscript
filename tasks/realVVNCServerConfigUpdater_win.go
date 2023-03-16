//go:build windows
// +build windows

package tasks

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/winreg"
)

const (
	DefaultWindowsExecPath = `C:\Program Files\RealVNC\VNC Server`
	DefaultWindowsExecName = `vncserver.exe`
)

var (
	HKLMBaseKey = `HKLM:\Software\RealVNC\vncserver`
	HKCUBaseKey = `HKCU:\Software\RealVNC\vncserver`
)

func (rvste *RealVNCServerTaskExecutor) applyConfigChanges(rvst *RealVNCServerTask) (addedCount int, updatedCount int, err error) {
	baseKey := getBaseKeyForServerMode(rvst.ServerMode)

	err = rvst.tracker.WithNewValues(func(fk string, fs FieldStatus) (err error) {
		fieldName := fs.Name
		regPath := fieldName
		regValue, err := rvst.getFieldValueAsString(fieldName)
		if err != nil {
			return err
		}

		updated := false
		desc := ""

		if fs.Clear {
			updated, desc, err = winreg.RemoveValue(baseKey, regPath)
			if err != nil {
				return err
			}
			if strings.Contains(desc, "removed") {
				updatedCount++
			}
		} else {
			updated, desc, err = winreg.SetValue(baseKey, regPath, regValue, winreg.REG_SZ)
			if err != nil {
				return err
			}
			if strings.Contains(desc, "added") {
				addedCount++
			} else if strings.Contains(desc, "updated") {
				updatedCount++
			}
		}

		if updated {
			err := rvst.tracker.SetChangeApplied(fk)
			if err != nil {
				return err
			}
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
