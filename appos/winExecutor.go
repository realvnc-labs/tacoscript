// +build windows

package appos

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

type OsExecutor struct {
}

func (oe OsExecutor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (oe OsExecutor) SetUser(userName, path string, cmd *exec.Cmd) error {
	logrus.Warn("user switch is not implemented in Windows")

	return nil
}
