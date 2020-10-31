// +build darwin

package pkg

import (
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"strings"
)

type OsPackageManagerCmdBuilder struct {}

func (ecb OsPackageManagerCmdBuilder) GetPkgManagerVersionCmd() string {
	return "brew --version"
}

func (ecb OsPackageManagerCmdBuilder) GetUpdateCmd(t *tasks.PkgTask) string {
	return "brew update"
}

func (ecb OsPackageManagerCmdBuilder) GetInstallCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()
	if t.Version != "" {
		for k, rawCmd := range rawCmds {
			rawCmd = fmt.Sprintf("%s@%s", rawCmd, t.Version)
			rawCmds[k] = rawCmd
		}
	}

	return []string{fmt.Sprintf("brew install %s", strings.Join(rawCmds, " "))}
}

func (ecb OsPackageManagerCmdBuilder) GetUninstallCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()

	return []string{fmt.Sprintf("brew uninstall %s", strings.Join(rawCmds, " "))}
}

func (ecb OsPackageManagerCmdBuilder) GetUpdateCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()

	return []string{fmt.Sprintf("brew upgrade %s", strings.Join(rawCmds, " "))}
}
