//go:build darwin
// +build darwin

package pkgmanager

import (
	"fmt"
	"strings"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

func BuildManagementCmdsProviders() ([]ManagementCmdsProvider, error) {
	return []ManagementCmdsProvider{
		OsPackageManagerCmdProvider{},
	}, nil
}

type OsPackageManagerCmdProvider struct{}

func (ecb OsPackageManagerCmdProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()
	rawInstallCmds := make([]string, 0, len(rawCmds))
	if t.Version != "" {
		for _, rawCmd := range rawCmds {
			rawInstallCmd := fmt.Sprintf("%s@%s", rawCmd, t.Version)
			rawInstallCmds = append(rawInstallCmds, rawInstallCmd)
		}
	} else {
		rawInstallCmds = rawCmds
	}

	return &ManagementCmds{
		VersionCmd:    "brew --version",
		UpgradeCmd:    "brew update",
		InstallCmds:   []string{fmt.Sprintf("brew install %s", strings.Join(rawInstallCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("brew uninstall %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("brew upgrade %s", strings.Join(rawCmds, " "))},
		ListCmd:       "brew list --formula --versions",
	}, nil
}
