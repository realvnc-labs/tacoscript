package exec

import (
	"context"
	"github.com/cloudradar-monitoring/tacoscript/conv"
	"io"
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
	Shell        string
}

type ShellParam struct {
	ShellName      string
	ShellPath      string
	ShellParams    []string
	RawShellString string
}

type CmdParam struct {
	Cmd          string
	Params       []string
	RawCmdString string
}
