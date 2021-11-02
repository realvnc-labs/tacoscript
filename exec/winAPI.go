//go:build windows
// +build windows

package exec

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

type OSApi struct {
}

func (oe OSApi) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (oe OSApi) SetUser(userName, path string, cmd *exec.Cmd) error {
	logrus.Warn("user switch is not implemented in Windows")

	return nil
}
