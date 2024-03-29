package pkgmanager

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/exec"
	"github.com/sirupsen/logrus"
)

type ManagementCmds struct {
	VersionCmd    string
	UpgradeCmd    string
	InstallCmds   []string
	UninstallCmds []string
	UpgradeCmds   []string
	ListCmd       string
	FilterFunc    func(ctx context.Context, rawPackages []string) []string
}

type ManagementCmdsProvider interface {
	GetManagementCmds(t *pkgtask.Task) (*ManagementCmds, error)
}

type PackageTaskManager struct {
	Runner                          exec.Runner
	ManagementCmdsProviderBuildFunc func() ([]ManagementCmdsProvider, error)
}

func (pm PackageTaskManager) ExecuteTask(ctx context.Context, t *pkgtask.Task) (res *pkgtask.ExecutionResult, err error) {
	managementCmdProviders, err := pm.ManagementCmdsProviderBuildFunc()
	if err != nil {
		return nil, err
	}

	if len(managementCmdProviders) == 0 {
		err = fmt.Errorf("no package manager providers for the current OS ")
		return
	}

	res = &pkgtask.ExecutionResult{}

	var managementCmds *ManagementCmds
	foundSupportedPackageManager := false
	triedCommands := make([]string, 0, len(managementCmdProviders))
	for _, managementCmdProvider := range managementCmdProviders {
		managementCmds, err = managementCmdProvider.GetManagementCmds(t)
		if err != nil {
			return nil, err
		}

		logrus.Debugf("will execute version command %s to check if package manager is installed", managementCmds.VersionCmd)

		versionResult := &pkgtask.ExecutionResult{}
		err = pm.run(ctx, t, versionResult, managementCmds.VersionCmd)
		if err != nil {
			triedCommands = append(triedCommands, fmt.Sprintf("%s: %v", managementCmds.VersionCmd, err))
			continue
		}

		logrus.Debugf(
			"version command '%s' success: %s, will use it for further package management",
			managementCmds.VersionCmd,
			versionResult.Output,
		)
		foundSupportedPackageManager = true
		break
	}

	if !foundSupportedPackageManager {
		return nil, fmt.Errorf(
			"cannot find a supported package manager on the host, tried package manager commands: %s",
			strings.Join(triedCommands, ", "),
		)
	}

	err = pm.updatePkgManagerIfNeeded(ctx, t, managementCmds)
	if err != nil {
		return nil, err
	}

	packagesListBefore, fetchPackagesErr := pm.getPackagesList(ctx, t, managementCmds)
	if fetchPackagesErr != nil {
		logrus.Warnf("failed to fetch packages list: %v", fetchPackagesErr)
	}

	err = pm.executePackageMethod(ctx, t, managementCmds, res)
	if err != nil {
		return nil, err
	}

	if fetchPackagesErr == nil {
		res.Changes = pm.getPackageDiff(ctx, managementCmds, t, packagesListBefore)
	}

	return res, nil
}

func (pm PackageTaskManager) executePackageMethod(
	ctx context.Context,
	t *pkgtask.Task,
	managementCmds *ManagementCmds,
	res *pkgtask.ExecutionResult,
) (err error) {
	switch t.ActionType {
	case pkgtask.ActionInstall:
		err = pm.installPackages(ctx, t, managementCmds, res)
		if err != nil {
			return err
		}
		res.Comment = fmt.Sprintf("The following packages are ensured to be installed: %s", pm.getAffectedPackagesStr(t))
	case pkgtask.ActionUninstall:
		err = pm.uninstallPackages(ctx, t, managementCmds, res)
		if err != nil {
			return err
		}
		res.Comment = fmt.Sprintf("The following packages are ensured to be uninstalled: %s", pm.getAffectedPackagesStr(t))
	case pkgtask.ActionUpdate:
		err = pm.updatePackages(ctx, t, managementCmds, res)
		if err != nil {
			return err
		}
		res.Comment = fmt.Sprintf("The following packages are ensured to be updated: %s", pm.getAffectedPackagesStr(t))
	default:
		return fmt.Errorf("unknown action type '%v' for task %s", t.ActionType, t.TypeName)
	}

	return nil
}

func (pm PackageTaskManager) getPackageDiff(
	ctx context.Context,
	managementCmds *ManagementCmds,
	t *pkgtask.Task,
	packagesListBefore []string,
) map[string]string {
	res := map[string]string{}
	packagesListAfter, err := pm.getPackagesList(ctx, t, managementCmds)
	if err != nil {
		logrus.Warnf("failed to fetch packages list: %v", err)
	}

	packagesDiff := CalcDiff(packagesListBefore, packagesListAfter)
	if packagesDiff != nil {
		if len(packagesDiff.Added) > 0 {
			for k, addedPkg := range packagesDiff.Added {
				res[fmt.Sprintf("added [%d]", k)] = addedPkg
			}
		}
		if len(packagesDiff.Removed) > 0 {
			for k, removedPkg := range packagesDiff.Removed {
				res[fmt.Sprintf("removed [%d]", k)] = removedPkg
			}
		}
	}

	return res
}

func (pm PackageTaskManager) getAffectedPackagesStr(t *pkgtask.Task) string {
	packages := make([]string, 0, len(t.Named.Names)+1)
	if t.Named.Name != "" {
		packages = append(packages, t.Named.Name)
	}

	for _, pkg := range t.Named.Names {
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return strings.Join(packages, ", ")
}

func (pm PackageTaskManager) installPackages(
	ctx context.Context,
	t *pkgtask.Task,
	mngtCmds *ManagementCmds,
	res *pkgtask.ExecutionResult,
) (err error) {
	logrus.Debugf("will install packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.InstallCmds))

	err = pm.run(ctx, t, res, mngtCmds.InstallCmds...)

	return err
}

func (pm PackageTaskManager) uninstallPackages(
	ctx context.Context,
	t *pkgtask.Task,
	mngtCmds *ManagementCmds,
	res *pkgtask.ExecutionResult,
) (err error) {
	logrus.Debugf("will uninstall packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.UninstallCmds))

	err = pm.run(ctx, t, res, mngtCmds.UninstallCmds...)

	return err
}

func (pm PackageTaskManager) updatePackages(
	ctx context.Context,
	t *pkgtask.Task,
	mngtCmds *ManagementCmds,
	res *pkgtask.ExecutionResult,
) (err error) {
	logrus.Debugf("will upgrade packages by executing %s", conv.ConvertSourceToJSONStrIfPossible(mngtCmds.UpgradeCmds))

	err = pm.run(ctx, t, res, mngtCmds.UpgradeCmds...)

	return err
}

func (pm PackageTaskManager) updatePkgManagerIfNeeded(
	ctx context.Context,
	t *pkgtask.Task,
	mngtCmds *ManagementCmds,
) (err error) {
	if !t.ShouldRefresh {
		return
	}

	res := &pkgtask.ExecutionResult{}
	logrus.Debugf("will update package manager: %s", mngtCmds.UpgradeCmd)
	err = pm.run(ctx, t, res, mngtCmds.UpgradeCmd)
	if err != nil {
		return err
	}

	logrus.Debugf("update result: %s", res.Output)
	return nil
}

func (pm PackageTaskManager) getPackagesList(
	ctx context.Context,
	t *pkgtask.Task,
	mngtCmds *ManagementCmds,
) (packagesList []string, err error) {
	logrus.Debugf("will fetch packages list: %s", mngtCmds.ListCmd)

	res := &pkgtask.ExecutionResult{}

	err = pm.run(ctx, t, res, mngtCmds.ListCmd)
	if err != nil {
		return nil, err
	}

	packagesList = strings.Split(res.Output, utils.LineBreak)
	logrus.Debugf("got %d packages", len(packagesList))

	if mngtCmds.FilterFunc != nil {
		packagesList = mngtCmds.FilterFunc(ctx, packagesList)
	}

	return packagesList, nil
}

func (pm PackageTaskManager) run(
	ctx context.Context,
	t *pkgtask.Task,
	res *pkgtask.ExecutionResult,
	rawCmds ...string,
) (err error) {
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

	if err != nil {
		return err
	}

	logrus.Debugf("Cmds %s success", conv.ConvertSourceToJSONStrIfPossible(rawCmds))

	logrus.Debugf(
		"stdOut: %s, stdErr: %s",
		stderrBuf.String(),
		stdoutBuf.String(),
	)

	res.Output = stderrBuf.String() + stdoutBuf.String()
	res.Pid = execCtx.Pid

	return nil
}
