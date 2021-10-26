//go:build darwin
// +build darwin

package pkg

import (
	"fmt"
	"strings"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

func BuildManagementCmdsProviders(t *tasks.PkgTask) ([]ManagementCmdsProvider, error) {
	return []ManagementCmdsProvider{
		BrewPackageManagerCmdProvider{},
	}, nil
}

type BrewPackageManagerCmdProvider struct{}

func (ecb BrewPackageManagerCmdProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
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

func (ecb BrewPackageManagerCmdProvider) GetName() string {
	return tasks.BrewManager
}
