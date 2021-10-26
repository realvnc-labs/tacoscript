// +build windows

package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type ChocoPackageManagerCmdProvider struct{}

func BuildManagementCmdsProviders(t *tasks.PkgTask) ([]ManagementCmdsProvider, error) {
	if t.Manager == tasks.WingetManager {
		return []ManagementCmdsProvider{
			WingetPackageManagerCmdProvider{},
		}, nil
	}

	return []ManagementCmdsProvider{
		ChocoPackageManagerCmdProvider{},
	}, nil
}

func (ecb ChocoPackageManagerCmdProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()

	versionStr := ""
	if t.Version != "" {
		versionStr += " " + t.Version
	}

	return &ManagementCmds{
		VersionCmd:    "choco --version",
		UpgradeCmd:    "choco upgrade -y chocolatey",
		InstallCmds:   []string{fmt.Sprintf("choco install -y %s%s", strings.Join(rawCmds, " "), versionStr)},
		UninstallCmds: []string{fmt.Sprintf("choco uninstall -y %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("choco upgrade -y %s", strings.Join(rawCmds, " "))},
		ListCmd:       "choco list --local-only",
		FilterFunc: func(ctx context.Context, rawPackages []string) []string {
			res := make([]string, 0, len(rawPackages))
			for _, rawPackage := range rawPackages {
				if strings.Contains(rawPackage, "packages installed") {
					continue
				}
				res = append(res, rawPackage)
			}

			return res
		},
	}, nil
}

func (ecb ChocoPackageManagerCmdProvider) GetName() string {
	return tasks.ChocoManager
}

type WingetPackageManagerCmdProvider struct{}

func (ecb WingetPackageManagerCmdProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()

	versionStr := ""
	if t.Version != "" {
		versionStr += " " + t.Version
	}

	return &ManagementCmds{
		VersionCmd:    "winget --version",
		UpgradeCmd:    "winget source update",
		InstallCmds:   []string{fmt.Sprintf("winget install -e -h --accept-package-agreements --accept-source-agreements %s%s", strings.Join(rawCmds, " "), versionStr)},
		UninstallCmds: []string{fmt.Sprintf("winget uninstall -e -h --accept-source-agreements %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("winget upgrade -e -h --accept-package-agreements --accept-source-agreements %s", strings.Join(rawCmds, " "))},
		ListCmd:       "winget list --accept-source-agreements",
		FilterFunc: func(ctx context.Context, rawPackages []string) []string {
			res := make([]string, 0, len(rawPackages))
			for i, rawPackage := range rawPackages {
				if i < 2 || rawPackage == "" {
					continue
				}
				res = append(res, rawPackage)
			}

			return res
		},
	}, nil
}

func (ecb WingetPackageManagerCmdProvider) GetName() string {
	return tasks.WingetManager
}
