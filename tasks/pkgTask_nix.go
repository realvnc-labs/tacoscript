// +build linux

package tasks

const (
	AptManager    = "apt"
	YumManager    = "yum"
	DnfManager    = "dnf"
	AptGetManager = "apt-get"
)

var supportedManagers = map[string]string{
	AptManager:    AptManager,
	YumManager:    YumManager,
	DnfManager:    DnfManager,
	AptGetManager: AptGetManager,
}
