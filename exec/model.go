package exec

import (
	"context"
	"io"

	"github.com/realvnc-labs/tacoscript/conv"
)

type Context struct {
	Ctx          context.Context
	StdoutWriter io.Writer
	StderrWriter io.Writer
	WorkingDir   string
	User         string
	Path         string
	Envs         conv.KeyValues
	Cmds         []string
	Pid          int
	Shell        string
}

func (c *Context) Copy() Context {
	return Context{
		Ctx:          c.Ctx,
		StdoutWriter: c.StdoutWriter,
		StderrWriter: c.StderrWriter,
		WorkingDir:   c.WorkingDir,
		User:         c.User,
		Path:         c.Path,
		Envs:         c.Envs,
		Cmds:         c.Cmds,
		Shell:        c.Shell,
	}
}

type ShellParam struct {
	ShellName      string
	ShellPath      string
	ShellParams    []string
	RawShellString string
}

type RunError struct {
	Err      error
	ExitCode int
}

func (re RunError) Error() string {
	return re.Err.Error()
}
