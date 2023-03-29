package exec

import (
	"io"
	"os/exec"
)

type SystemAPIMock struct {
	StdOutText string
	StdErrText string
	Cmds       []*exec.Cmd
	ErrToGive  error

	UserNameInput      string
	UserNamePathInput  string
	UserSetErrToReturn error
	Callback           func(cmd *exec.Cmd) error
}

func (oem *SystemAPIMock) Run(cmd *exec.Cmd) error {
	oem.Cmds = append(oem.Cmds, cmd)

	if oem.Callback != nil {
		return oem.Callback(cmd)
	}

	if oem.StdOutText != "" && cmd.Stdout != nil {
		_, err := cmd.Stdout.Write([]byte(oem.StdOutText))
		if err != nil {
			return err
		}
	}

	if oem.StdErrText != "" && cmd.Stderr != nil {
		_, err := cmd.Stderr.Write([]byte(oem.StdErrText))
		if err != nil {
			return err
		}
	}

	return oem.ErrToGive
}

func (oem *SystemAPIMock) SetUser(userName, path string, cmd *exec.Cmd) error {
	oem.UserNameInput = userName
	oem.UserNamePathInput = path

	return oem.UserSetErrToReturn
}

type RunnerMock struct {
	GivenExecContexts []*Context
	ErrToReturn       error
	RunOutputCallback func(stdOutWriter, stdErrWriter io.Writer)
}

func (rm *RunnerMock) Run(execContext *Context) error {
	rm.GivenExecContexts = append(rm.GivenExecContexts, execContext)
	if rm.RunOutputCallback != nil {
		rm.RunOutputCallback(execContext.StdoutWriter, execContext.StderrWriter)
	}
	return rm.ErrToReturn
}
