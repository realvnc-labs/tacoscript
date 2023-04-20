//go:build windows
// +build windows

package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
)

type OsPackageManagerCmdProvider struct{}

func BuildManagementCmdsProviders() ([]ManagementCmdsProvider, error) {
	return []ManagementCmdsProvider{
		OsPackageManagerCmdProvider{},
	}, nil
}

func (ecb OsPackageManagerCmdProvider) GetManagementCmds(t *pkgtask.Task) (*ManagementCmds, error) {
	rawCmds := t.Named.GetNames()

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
