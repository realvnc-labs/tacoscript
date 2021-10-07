package tasks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/conv"
	"gopkg.in/yaml.v2"

	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"

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

type pkgContextProc func(t *PkgTask, path string, val interface{}) error

var pkgContextProcMap = map[string]pkgContextProc{
	NameField: func(t *PkgTask, path string, val interface{}) error {
		t.Name = fmt.Sprint(val)
		return nil
	},
	ShellField: func(t *PkgTask, path string, val interface{}) error {
		t.Shell = fmt.Sprint(val)
		return nil
	},
	RequireField: func(t *PkgTask, path string, val interface{}) error {
		var err error
		t.Require, err = parseRequireField(val, path)
		return err
	},
	OnlyIf: func(t *PkgTask, path string, val interface{}) error {
		var err error
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	},
	Unless: func(t *PkgTask, path string, val interface{}) error {
		var err error
		t.Unless, err = parseUnlessField(val, path)
		return err
	},
	Version: func(t *PkgTask, path string, val interface{}) error {
		t.Version = fmt.Sprint(val)
		return nil
	},
	Refresh: func(t *PkgTask, path string, val interface{}) error {
		t.ShouldRefresh = parseBoolField(val)
		return nil
	},
	NamesField: func(t *PkgTask, path string, val interface{}) error {
		var names []string
		var err error
		names, err = conv.ConvertToValues(val, path)
		t.Names = names
		return err
	},
}

func (fmtb PkgTaskBuilder) Build(typeName, path string, ctx interface{}) (Task, error) {
	t := &PkgTask{
		TypeName: typeName,
		Path:     path,
	}

	switch typeName {
	case PkgInstalled:
		t.ActionType = ActionInstall
	case PkgRemoved:
		t.ActionType = ActionUninstall
	case PkgUpgraded:
		t.ActionType = ActionUpdate
	}

	errs := &utils.Errors{}

	for _, item := range ctx.([]interface{}) {
		row := item.(yaml.MapSlice)[0]
		key := row.Key.(string)
		val := row.Value
		f, ok := pkgContextProcMap[key]
		if !ok {
			continue
		}
		errs.Add(f(t, path, val))
	}

	return t, errs.ToError()
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
	OnlyIf        []string
	Unless        []string
}

func (pt *PkgTask) GetName() string {
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

type PackageManager interface {
	ExecuteTask(ctx context.Context, t *PkgTask) (output string, err error)
}

type PkgTaskExecutor struct {
	PackageManager PackageManager
	Runner         exec2.Runner
}

func (pte *PkgTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := ExecutionResult{}

	pkgTask, ok := task.(*PkgTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to PkgTask", task)
		return execRes
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec2.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
	}
	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := pte.shouldBeExecuted(execCtx, pkgTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if skipReason != "" {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		execRes.SkipReason = skipReason
		return execRes
	}

	start := time.Now()

	output, err := pte.PackageManager.ExecuteTask(ctx, pkgTask)
	execRes.Err = err
	execRes.StdOut = output
	execRes.IsSkipped = false
	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", task.GetPath(), execRes.Duration)
	return execRes
}

func (pte *PkgTaskExecutor) checkOnlyIfs(ctx *exec2.Context, pkgTask *PkgTask) (isSuccess bool, err error) {
	if len(pkgTask.OnlyIf) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = pkgTask.OnlyIf
	err = pte.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Debugf("will skip %s since onlyif condition has failed: %v", pkgTask, runErr)
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (pte *PkgTaskExecutor) shouldBeExecuted(
	ctx *exec2.Context,
	pkgTask *PkgTask,
) (skipReason string, err error) {
	isSuccess, err := pte.checkOnlyIfs(ctx, pkgTask)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return onlyIfConditionFailedReason, nil
	}

	isExpectationSuccess, err := pte.checkUnless(ctx, pkgTask)
	if err != nil {
		return "", err
	}

	if !isExpectationSuccess {
		skipReason = fmt.Sprintf("unless condition is true, will skip %s", pkgTask.Path)
		logrus.Debug(skipReason)
		return skipReason, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", pkgTask)
	return "", nil
}

func (pte *PkgTaskExecutor) checkUnless(ctx *exec2.Context, pkgTask *PkgTask) (isExpectationSuccess bool, err error) {
	if len(pkgTask.Unless) == 0 {
		isExpectationSuccess = true
		err = nil
		return
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = pkgTask.Unless

	err = pte.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Infof("will continue cmd since at least one unless condition has failed: %v", runErr)
			return true, nil
		}

		return false, err
	}

	logrus.Infof("any unless condition didn't fail for task '%s'", pkgTask.Path)
	return false, nil
}
