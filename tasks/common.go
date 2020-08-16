package tasks

import (
	"context"
	"time"
)

var cParamShells = map [string]string{
	"zsh": "-c",
	"bash": "-c",
	"sh": "-c",
	"cmd.exe": "/C",
}

type Scripts []Script

type Script struct {
	ID    string
	Tasks []Task
}

type Task interface {
	GetName() string
	Execute(ctx context.Context) ExecutionResult
	Validate() error
	GetPath() string
	GetRequirements() []string
}

type ExecutionResult struct {
	Err       error
	Duration  time.Duration
	StdErr    string
	StdOut    string
	IsSkipped bool
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
