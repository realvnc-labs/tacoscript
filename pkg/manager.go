package pkg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/conv"
	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/sirupsen/logrus"
)

type ManagementCmds struct {
	VersionCmd    string
	UpgradeCmd    string
	InstallCmds   []string
	UninstallCmds []string
	UpgradeCmds   []string
}

type ManagementCmdsProvider interface {
	GetManagementCmds(t *tasks.PkgTask) (ManagementCmds, error)
}

type PackageTaskManager struct {
	Runner                     exec.Runner
	PackageManagerCmdProviders []ManagementCmdsProvider
}

func (pm PackageTaskManager) ExecuteTask(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	if len(pm.PackageManagerCmdProviders) == 0 {
		err = fmt.Errorf("no package manager providers for the current OS ")
		return
	}

	var managementCmds ManagementCmds
	for _, managementCmdProvider := range pm.PackageManagerCmdProviders {
		managementCmds, err = managementCmdProvider.GetManagementCmds(t)
		if err != nil {
			return "", err
		}

		logrus.Debugf("will execute version command %s to check if package manager is installed", managementCmds.VersionCmd)

		output, err = pm.run(ctx, t, managementCmds.VersionCmd)
		if err == nil {
			logrus.Debugf("version command success: %s, will use it for further package management", managementCmds.VersionCmd)
			break
		}
	}
	if err != nil {
		return
	}

	output, err = pm.updatePkgManagerIfNeeded(ctx, t, managementCmds)
	if err != nil {
		return
	}

	switch t.ActionType {
	case tasks.ActionInstall:
		output, err = pm.installPackages(ctx, t, managementCmds)
		return
	case tasks.ActionUninstall:
		output, err = pm.uninstallPackages(ctx, t, managementCmds)
		return
	case tasks.ActionUpdate:
		output, err = pm.updatePackages(ctx, t, managementCmds)
		return
	default:
		err = fmt.Errorf("unknown action type '%v' for task %s", t.ActionType, t.TypeName)
		return
	}
}

func (pm PackageTaskManager) installPackages(ctx context.Context, t *tasks.PkgTask, mngtCmds ManagementCmds) (output string, err error) {
	logrus.Debugf("will install packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.InstallCmds))

	output, err = pm.run(ctx, t, mngtCmds.InstallCmds...)

	return
}

func (pm PackageTaskManager) uninstallPackages(ctx context.Context, t *tasks.PkgTask, mngtCmds ManagementCmds) (output string, err error) {
	logrus.Debugf("will uninstall packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.UninstallCmds))

	output, err = pm.run(ctx, t, mngtCmds.UninstallCmds...)

	return
}

func (pm PackageTaskManager) updatePackages(ctx context.Context, t *tasks.PkgTask, mngtCmds ManagementCmds) (output string, err error) {
	logrus.Debugf("will upgrade packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.UpgradeCmds))

	output, err = pm.run(ctx, t, mngtCmds.UpgradeCmds...)

	return
}

func (pm PackageTaskManager) updatePkgManagerIfNeeded(ctx context.Context, t *tasks.PkgTask, mngtCmds ManagementCmds) (output string, err error) {
	if !t.ShouldRefresh {
		return
	}

	logrus.Debugf("will update package manager: %s", mngtCmds.UpgradeCmd)
	output, err = pm.run(ctx, t, mngtCmds.UpgradeCmd)

	return
}

func (pm PackageTaskManager) run(ctx context.Context, t *tasks.PkgTask, rawCmds ...string) (output string, err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		Path:         t.Path,
		Cmds:         rawCmds,
		Shell:        t.Shell,
	}

	err = pm.Runner.Run(execCtx)

	if err == nil {
		logrus.Debugf("Cmds %s success", conv.ConvertSourceToJSONStrIfPossible(rawCmds))
	}

	logrus.Debugf(
		"stdOut: %s, stdErr: %s",
		stderrBuf.String(),
		stdoutBuf.String(),
	)

	output = stderrBuf.String() + stdoutBuf.String()

	return
}
