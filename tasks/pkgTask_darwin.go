//go:build darwin
// +build darwin

package tasks

const (
	BrewManager  = "brew"
)

var supportedManagers = map[string]string{BrewManager: BrewManager}
