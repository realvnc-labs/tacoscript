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

type PackageManagerCmdBuilder interface {
	GetPkgManagerVersionCmd() string
	GetUpdateCmd(t *tasks.PkgTask) string
	GetInstallCmdForPackages(t *tasks.PkgTask) []string
	GetUninstallCmdForPackages(t *tasks.PkgTask) []string
	GetUpdateCmdForPackages(t *tasks.PkgTask) []string
}

type PackageTaskManager struct {
	Runner                   exec.Runner
	PackageManagerCmdBuilder PackageManagerCmdBuilder
}

func (pm PackageTaskManager) ExecuteTask(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	output, err = pm.checkIfPkgManagerExists(ctx, t)
	if err != nil {
		return
	}

	output, err = pm.updatePkgManagerIfNeeded(ctx, t)
	if err != nil {
		return
	}

	switch t.ActionType {
	case tasks.ActionInstall:
		output, err = pm.installPackages(ctx, t)
		return
	case tasks.ActionUninstall:
		output, err = pm.uninstallPackages(ctx, t)
		return
	case tasks.ActionUpdate:
		output, err = pm.updatePackages(ctx, t)
		return
	default:
		err = fmt.Errorf("unknown action type '%v' for task %s", t.ActionType, t.TypeName)
		return
	}
}

func (pm PackageTaskManager) installPackages(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	rawCmds := pm.PackageManagerCmdBuilder.GetInstallCmdForPackages(t)

	logrus.Debugf("will install packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(rawCmds))

	output, err = pm.run(ctx, t, rawCmds...)

	return
}

func (pm PackageTaskManager) uninstallPackages(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	rawCmds := pm.PackageManagerCmdBuilder.GetUninstallCmdForPackages(t)

	logrus.Debugf("will uninstall packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(rawCmds))

	output, err = pm.run(ctx, t, rawCmds...)

	return
}

func (pm PackageTaskManager) updatePackages(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	rawCmds := pm.PackageManagerCmdBuilder.GetUpdateCmdForPackages(t)

	logrus.Debugf("will update packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(rawCmds))

	output, err = pm.run(ctx, t, rawCmds...)

	return
}

func (pm PackageTaskManager) checkIfPkgManagerExists(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	versionCmd := pm.PackageManagerCmdBuilder.GetPkgManagerVersionCmd()

	logrus.Debugf("will execute version command %s to check if package manager is installed", versionCmd)

	output, err = pm.run(ctx, t, versionCmd)

	return
}

func (pm PackageTaskManager) updatePkgManagerIfNeeded(ctx context.Context, t *tasks.PkgTask) (output string, err error) {
	if !t.ShouldRefresh {
		return
	}

	pkgUpdateCmd := pm.PackageManagerCmdBuilder.GetUpdateCmd(t)
	logrus.Debugf("will update package manager: %s", pkgUpdateCmd)
	output, err = pm.run(ctx, t, pkgUpdateCmd)

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
		logrus.Debugf("Cmds %s success",conv.ConvertSourceToJSONStrIfPossible(rawCmds))
	}

	logrus.Debugf(
		"stdOut: %s, stdErr: %s",
		stderrBuf.String(),
		stdoutBuf.String(),
	)

	output = stderrBuf.String() + stdoutBuf.String()

	return
}
