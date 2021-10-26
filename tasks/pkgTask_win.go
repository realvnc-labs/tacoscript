// +build windows

package tasks

const (
	ChocoManager  = "choco"
	WingetManager = "winget"
)

var supportedManagers = map[string]string{ChocoManager: ChocoManager, WingetManager: WingetManager}
