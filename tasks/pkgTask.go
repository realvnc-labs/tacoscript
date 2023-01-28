package tasks

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

type PkgActionType int

const (
	ActionInstall PkgActionType = iota + 1
	ActionUninstall
	ActionUpdate
)

type PkgTaskBuilder struct {
}

var pkgTaskParamsFnMap = taskParamsFnMap{
	NameField: func(task Task, path string, val interface{}) error {
		t := task.(*PkgTask)
		t.Name = fmt.Sprint(val)
		return nil
	},
	ShellField: func(task Task, path string, val interface{}) error {
		t := task.(*PkgTask)
		t.Shell = fmt.Sprint(val)
		return nil
	},
	RequireField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*PkgTask)
		t.Require, err = parseRequireField(val, path)
		return err
	},
	CreatesField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*PkgTask)
		t.Creates, err = parseCreatesField(val, path)
		return err
	},
	OnlyIfField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*PkgTask)
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	},
	UnlessField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*PkgTask)
		t.Unless, err = parseUnlessField(val, path)
		return err
	},
	Version: func(task Task, path string, val interface{}) error {
		t := task.(*PkgTask)
		t.Version = fmt.Sprint(val)
		return nil
	},
	Refresh: func(task Task, path string, val interface{}) error {
		t := task.(*PkgTask)
		t.ShouldRefresh = parseBoolField(val)
		return nil
	},
	NamesField: func(task Task, path string, val interface{}) error {
		var names []string
		var err error
		t := task.(*PkgTask)
		names, err = conv.ConvertToValues(val, path)
		t.Names = names
		return err
	},
}

func (fmtb PkgTaskBuilder) Build(typeName, path string, params interface{}) (Task, error) {
	task := &PkgTask{
		TypeName: typeName,
		Path:     path,
	}

	switch typeName {
	case PkgInstalled:
		task.ActionType = ActionInstall
	case PkgRemoved:
		task.ActionType = ActionUninstall
	case PkgUpgraded:
		task.ActionType = ActionUpdate
	}

	errs := Build(typeName, path, params, task, pkgTaskParamsFnMap)

	return task, errs.ToError()
}

type PkgTask struct {
	ActionType PkgActionType
	TypeName   string
	Path       string
	NamedTask
	Shell         string
	Version       string
	ShouldRefresh bool
	Require       []string
	Creates       []string
	OnlyIf        []string
	Unless        []string
}

func (pt *PkgTask) GetTypeName() string {
	return pt.TypeName
}

func (pt *PkgTask) GetRequirements() []string {
	return pt.Require
}

func (pt *PkgTask) Validate() error {
	errs := &utils.Errors{}

	err1 := ValidateRequired(pt.Name, pt.Path+"."+NameField)
	err2 := ValidateRequiredMany(pt.Names, pt.Path+"."+NamesField)

	if err1 != nil && err2 != nil {
		errs.Add(err1)
		errs.Add(err2)
	}

	if pt.ActionType == 0 {
		errs.Add(fmt.Errorf("unknown pkg task type: %s", pt.TypeName))
	}

	return errs.ToError()
}

func (pt *PkgTask) GetPath() string {
	return pt.Path
}

func (pt *PkgTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", pt.TypeName, pt.GetPath())
}

func (pt *PkgTask) GetOnlyIfCmds() []string {
	return pt.OnlyIf
}

func (pt *PkgTask) GetUnlessCmds() []string {
	return pt.Unless
}

func (pt *PkgTask) GetCreatesFilesList() []string {
	return pt.Creates
}

type PackageManagerExecutionResult struct {
	Output  string
	Comment string
	Changes map[string]string
	Pid     int
}

type PackageManager interface {
	ExecuteTask(ctx context.Context, t *PkgTask) (res *PackageManagerExecutionResult, err error)
}

type PkgTaskExecutor struct {
	PackageManager PackageManager
	Runner         tacoexec.Runner
	FsManager      *utils.FsManager
}

func (pte *PkgTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := ExecutionResult{}

	pkgTask, ok := task.(*PkgTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to PkgTask", task)
		return execRes
	}

	execRes.Name = strings.Join(pkgTask.GetNames(), "; ")

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := shouldCheckConditionals(execCtx, pte.FsManager, pte.Runner, pkgTask)
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

	logrus.Debugf("the task '%s' is finished for %v", execRes.Name, execRes.Duration)
	return execRes
}
