package cmdrun

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/shared/executionresult"
	"github.com/realvnc-labs/tacoscript/tasks/shared/names"

	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/realvnc-labs/tacoscript/conv"

	"github.com/sirupsen/logrus"
)

const (
	TaskType = "cmd.run"
)

type Task struct {
	TypeName string
	Path     string
	Named    names.TaskNames
	Envs     conv.KeyValues

	WorkingDir string   `taco:"cwd"`
	User       string   `taco:"user"`
	Shell      string   `taco:"shell"`
	Creates    []string `taco:"creates"`
	Require    []string `taco:"require"`
	OnlyIf     []string `taco:"onlyif"`
	Unless     []string `taco:"unless"`

	// aborts task execution if one task fails
	AbortOnError bool
}

func (crt *Task) GetTypeName() string {
	return crt.TypeName
}

func (crt *Task) GetRequirements() []string {
	return crt.Require
}

func (crt *Task) Validate(goos string) error {
	errs := &utils.Errors{}
	err1 := tasks.ValidateRequired(crt.Named.Name, crt.Path+"."+tasks.NameField)
	err2 := tasks.ValidateRequiredMany(crt.Named.Names, crt.Path+"."+tasks.NamesField)

	if err1 != nil && err2 != nil {
		errs.Add(err1)
		errs.Add(err2)
		return errs.ToError()
	}

	return nil
}

func (crt *Task) GetPath() string {
	return crt.Path
}

func (crt *Task) String() string {
	return conv.ConvertSourceToJSONStrIfPossible(crt)
}

func (crt *Task) GetOnlyIfCmds() []string {
	return crt.OnlyIf
}

func (crt *Task) GetUnlessCmds() []string {
	return crt.Unless
}

func (crt *Task) GetCreatesFilesList() []string {
	return crt.Creates
}

type CrtExecutor struct {
	Runner    tacoexec.Runner
	FsManager tasks.FsManager
}

func (crte *CrtExecutor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	execRes := executionresult.ExecutionResult{}
	cmdRunTask, ok := task.(*Task)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to Task", task)
		return execRes
	}
	execRes.Name = strings.Join(cmdRunTask.Named.GetNames(), "; ")

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   cmdRunTask.WorkingDir,
		User:         cmdRunTask.User,
		Path:         cmdRunTask.Path,
		Envs:         cmdRunTask.Envs,
		Cmds:         cmdRunTask.Named.GetNames(),
		Shell:        cmdRunTask.Shell,
	}

	shouldNotBeExecutedReason, err := tasks.CheckConditionals(execCtx, crte.FsManager, crte.Runner, cmdRunTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if shouldNotBeExecutedReason != "" {
		execRes.IsSkipped = true
		execRes.SkipReason = shouldNotBeExecutedReason
		execRes.Comment = `Command "` + execRes.Name + `" did not run: ` + shouldNotBeExecutedReason
		return execRes
	}

	start := time.Now()

	err = crte.Runner.Run(execCtx)

	execRes.Duration = time.Since(start)
	if err != nil {
		execRes.Err = err
	}
	logrus.Debugf("execution of %s has finished, took: %v", cmdRunTask.Named.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()
	execRes.Pid = execCtx.Pid

	return execRes
}
