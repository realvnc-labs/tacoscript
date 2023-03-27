package pkgtask

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
	"github.com/realvnc-labs/tacoscript/tasks/namedtask"

	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

type PkgActionType int

const (
	TaskTypePkgInstalled = "pkg.installed"
	TaskTypePkgRemoved   = "pkg.removed"
	TaskTypePkgUpgraded  = "pkg.uptodate"

	ActionInstall PkgActionType = iota + 1
	ActionUninstall
	ActionUpdate
)

type PTask struct {
	ActionType PkgActionType
	TypeName   string
	Path       string
	Named      namedtask.NamedTask

	Shell         string   `taco:"shell"`
	Version       string   `taco:"version"`
	ShouldRefresh bool     `taco:"refresh"`
	Require       []string `taco:"require"`
	Creates       []string `taco:"creates"`
	OnlyIf        []string `taco:"onlyif"`
	Unless        []string `taco:"unless"`

	Updated bool
}

func (pt *PTask) GetTypeName() string {
	return pt.TypeName
}

func (pt *PTask) GetRequirements() []string {
	return pt.Require
}

func (pt *PTask) Validate(goos string) error {
	errs := &utils.Errors{}

	err1 := tasks.ValidateRequired(pt.Named.Name, pt.Path+"."+tasks.NameField)
	err2 := tasks.ValidateRequiredMany(pt.Named.Names, pt.Path+"."+tasks.NamesField)

	if err1 != nil && err2 != nil {
		errs.Add(err1)
		errs.Add(err2)
	}

	if pt.ActionType == 0 {
		errs.Add(fmt.Errorf("unknown pkg task type: %s", pt.TypeName))
	}

	return errs.ToError()
}

func (pt *PTask) GetPath() string {
	return pt.Path
}

func (pt *PTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", pt.TypeName, pt.GetPath())
}

func (pt *PTask) GetOnlyIfCmds() []string {
	return pt.OnlyIf
}

func (pt *PTask) GetUnlessCmds() []string {
	return pt.Unless
}

func (pt *PTask) GetCreatesFilesList() []string {
	return pt.Creates
}

type PackageManagerExecutionResult struct {
	Output  string
	Comment string
	Changes map[string]string
	Pid     int
}

type PackageManager interface {
	ExecuteTask(ctx context.Context, t *PTask) (res *PackageManagerExecutionResult, err error)
}

type PtExecutor struct {
	PackageManager PackageManager
	Runner         tacoexec.Runner
	FsManager      *utils.FsManager
}

func (pte *PtExecutor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := executionresult.ExecutionResult{}

	pkgTask, ok := task.(*PTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to PTask", task)
		return execRes
	}

	execRes.Name = strings.Join(pkgTask.Named.GetNames(), "; ")

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := tasks.CheckConditionals(execCtx, pte.FsManager, pte.Runner, pkgTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if skipReason != "" {
		logrus.Debugf("the task '%s' will be be skipped", execRes.Name)
		execRes.IsSkipped = true
		execRes.SkipReason = skipReason
		return execRes
	}

	start := time.Now()

	pkgExecResult, err := pte.PackageManager.ExecuteTask(ctx, pkgTask)
	execRes.Err = err
	if pkgExecResult != nil {
		execRes.StdOut = pkgExecResult.Output
		execRes.Comment = pkgExecResult.Comment
		execRes.Changes = pkgExecResult.Changes
		execRes.Pid = pkgExecResult.Pid
	}

	execRes.IsSkipped = false
	execRes.Duration = time.Since(start)

	pkgTask.Updated = true

	logrus.Debugf("the task '%s' is finished for %v", execRes.Name, execRes.Duration)
	return execRes
}
