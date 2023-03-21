//go:build linux
// +build linux

package pkgmanager

import (
	"fmt"
	"strings"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/sirupsen/logrus"
)

var osPlatform string

var linuxOSBuilderMap = map[string][]ManagementCmdsProvider{
	"ubuntu": {
		AptCmdsProvider{},
		AptGetCmdsProvider{},
	},
	"debian": {
		AptCmdsProvider{},
		AptGetCmdsProvider{},
	},
	"centos": {
		DnfCmdsProvider{},
		YumCmdsProvider{},
	},
	"fedora": {
		DnfCmdsProvider{},
		YumCmdsProvider{},
	},
}

func init() {
	osDataProvider := utils.OSDataProvider{}
	templateVariables, err := osDataProvider.GetTemplateVariables()
	if err != nil {
		logrus.Error(err)
		return
	}
	osPlatform = templateVariables[utils.OSPlatform]
}

func BuildManagementCmdsProviders() ([]ManagementCmdsProvider, error) {
	linuxSpecificProviders, ok := linuxOSBuilderMap[osPlatform]
	if !ok {
		return []ManagementCmdsProvider{}, fmt.Errorf("unsupported linux version %s for package management commands", osPlatform)
	}

	return linuxSpecificProviders, nil
}

type AptCmdsProvider struct{}

func (ecb AptCmdsProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()
	rawInstallCmds := buildInstallCmds(rawCmds, t.Version)

	return &ManagementCmds{
		VersionCmd:    "apt --version",
		UpgradeCmd:    "apt update",
		InstallCmds:   []string{fmt.Sprintf("apt install -y %s", strings.Join(rawInstallCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("apt remove -y %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("apt upgrade -y %s", strings.Join(rawCmds, " "))},
		ListCmd:       "dpkg -l",
	}, nil
}

type AptGetCmdsProvider struct{}

func (ecb AptGetCmdsProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()
	rawInstallCmds := buildInstallCmds(rawCmds, t.Version)

	return &ManagementCmds{
		VersionCmd:    "apt-get --version",
		UpgradeCmd:    "apt-get update",
		InstallCmds:   []string{fmt.Sprintf("apt-get install -y %s", strings.Join(rawInstallCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("apt-get remove -y %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("apt-get upgrade -y %s", strings.Join(rawCmds, " "))},
		ListCmd:       "dpkg -l",
	}, nil
}

type YumCmdsProvider struct{}

func (ecb YumCmdsProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()
	rawInstallCmds := buildInstallCmds(rawCmds, t.Version)

	return &ManagementCmds{
		VersionCmd:    "yum --version",
		UpgradeCmd:    "yum update -y",
		InstallCmds:   []string{fmt.Sprintf("yum install -y %s", strings.Join(rawInstallCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("yum remove -y %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("yum upgrade -y %s", strings.Join(rawCmds, " "))},
		ListCmd:       "rpm -qa",
	}, nil
}

type DnfCmdsProvider struct{}

func (ecb DnfCmdsProvider) GetManagementCmds(t *tasks.PkgTask) (*ManagementCmds, error) {
	rawCmds := t.GetNames()
	rawInstallCmds := buildInstallCmds(rawCmds, t.Version)

	return &ManagementCmds{
		VersionCmd:    "dnf --version",
		UpgradeCmd:    "dnf update -y",
		InstallCmds:   []string{fmt.Sprintf("dnf install -y %s", strings.Join(rawInstallCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("dnf remove -y %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("dnf upgrade -y %s", strings.Join(rawCmds, " "))},
		ListCmd:       "rpm -qa",
	}, nil
}

func buildInstallCmds(rawCmds []string, version string) []string {
	rawInstallCmds := make([]string, 0, len(rawCmds))
	if version == "" {
		return rawCmds
	}
	for _, rawCmd := range rawCmds {
		rawInstallCmd := fmt.Sprintf("%s-%s", rawCmd, version)
		rawInstallCmds = append(rawInstallCmds, rawInstallCmd)
	}

	return rawInstallCmds
}
