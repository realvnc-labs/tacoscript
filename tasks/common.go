package tasks

import (
	"context"
	"errors"
	"strings"
	"time"
)

var cParamShells = []string{
	"zsh",
	"bash",
	"sh",
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
}

type ValidationErrors struct {
	Errs []error
}

func (ve *ValidationErrors) Add(err error) {
	if err == nil {
		return
	}

	ve.Errs = append(ve.Errs, err)
}

func (ve ValidationErrors) ToError() error {
	if len(ve.Errs) == 0 {
		return nil
	}

	rawErrors := make([]string, 0, len(ve.Errs))
	for _, err := range ve.Errs {
		rawErrors = append(rawErrors, err.Error())
	}

	return errors.New(strings.Join(rawErrors, ", "))
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
