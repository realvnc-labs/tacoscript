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
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/tasks/support/winregistry"
)

const (
	DefaultWindowsExec = `C:\Program Files\RealVNC\VNC Server\vncserver.exe`
)

var (
	HKLMBaseKey = `HKLM:\SOFTWARE\RealVNC\vncserver`
	HKCUBaseKey = `HKCU:\Software\RealVNC\vncserver`
	// TODO: (rs): Remove this when we re-introduce User and Virtual server modes.
	TestBaseKey = `HKCU:\Software\RealVNCTest\vncserver`
)

func (rvste *Executor) applyConfigChanges(rvst *Task) (addedCount int, updatedCount int, err error) {
	baseKey := getBaseKeyForServerMode(rvst.ServerMode)

	err = rvst.fieldTracker.WithNewValues(func(fieldName string, fs fieldstatus.FieldStatus) (err error) {
		regPath := fieldName
		regValue, err := rvst.getFieldValueAsString(fieldName)
		if err != nil {
			return err
		}

		desc := ""

		if fs.Clear {
			_, desc, err = winregistry.RemoveValue(baseKey, regPath)
			if err != nil {
				return err
			}
			if strings.Contains(desc, "removed") {
				updatedCount++
				logrus.Debugf(`removed key %s\%s`, baseKey, regPath)
			}
		} else {
			_, desc, err = winregistry.SetValue(baseKey, regPath, regValue, winregistry.REG_SZ)
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

		err = rvst.fieldTracker.SetChangeApplied(fieldName)
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

func (rvste *Executor) ReloadConfig(rvst *Task) (err error) {
	var cmd *exec.Cmd

	cmdLine := rvste.makeReloadPSCmdLine(rvst)

	cmd = exec.Command("powershell", cmdLine)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	cmdRunner := tacoexec.OSApi{}
	err = cmdRunner.Run(cmd)
	if err != nil {
		logrus.Debugf(`command output = %s`, outBuf.String())
		logrus.Debugf(`err output = %s`, errBuf.String())
		return fmt.Errorf("failed reloading vnc server configuration: %w", err)
	}

	logrus.Debugf(`config reloaded successfully`)

	return nil
}

func getBaseKeyForServerMode(serverMode string) (baseKey string) {
	baseKey = HKLMBaseKey
	if serverMode == UserServerMode {
		baseKey = HKCUBaseKey
	}
	if serverMode == TestServerMode {
		baseKey = TestBaseKey
	}
	return baseKey
}

// TODO: (rs): this should have unit test
func (rvste *Executor) makeReloadPSCmdLine(rvst *Task) (cmdLine string) {
	baseCmdLine := `Start-Process -FilePath '%s' -WindowStyle Hidden  -ArgumentList '%s'`
	argumentList := `service -reload`

	if rvst.ServerMode == UserServerMode {
		argumentList = `-reload`
	}

	cmd := DefaultWindowsExec
	if rvst.ReloadExecPath != "" {
		cmd = rvst.ReloadExecPath
		logrus.Debugf(`user specified reload_exec_path = %s`, cmd)
	}

	cmdLine = fmt.Sprintf(baseCmdLine, cmd, argumentList)
	return cmdLine
}
