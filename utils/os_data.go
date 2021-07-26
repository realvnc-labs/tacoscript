package utils

import (
	"strings"

	"github.com/shirou/gopsutil/host"
)

const (
	OSKernel = "taco_os_kernel" // windows, linux, freebsd
	// see https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-valid-goos-values
	OSFamily     = "taco_os_family"    // darwin, debian, redhat, debian, '', windows
	OSPlatform   = "taco_os_platform"  // darwin, ubuntu, centos, debian, alpine, windows
	OSName       = "taco_os_name"      // mac os x, ubuntu, centos linux, debian gnu/linux, alpine linux, windows server 2019 standard
	OSVersion    = "taco_os_version"   // 10.15.7, 20.04.1 LTS (Focal Fossa), 8 (Core), 10 (buster), '', 10.0
	Architecture = "taco_architecture" // x86_64
)

type OSDataProvider struct {
}

func (odp OSDataProvider) GetTemplateVariables() (map[string]interface{}, error) {
	h, err := host.Info()
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		OSKernel:     strings.ToLower(h.OS),
		OSFamily:     strings.ToLower(h.PlatformFamily),
		Architecture: strings.ToLower(h.KernelArch),
		OSPlatform:   strings.ToLower(h.Platform),
		OSName:       strings.ToLower(h.OS),
		OSVersion:    strings.ToLower(h.PlatformVersion),
	}, nil
}
