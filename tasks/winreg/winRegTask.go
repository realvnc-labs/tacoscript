package winreg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
	"github.com/realvnc-labs/tacoscript/winreg"

	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

type WinRegActionType int

const (
	TaskTypeWinRegPresent   = "win_reg.present"
	TaskTypeWinRegAbsent    = "win_reg.absent"
	TaskTypeWinRegAbsentKey = "win_reg.absent_key"

	ActionWinRegPresent WinRegActionType = iota + 1
	ActionWinRegAbsent
	ActionWinRegAbsentKey
)

var ErrUnknownWinRegAction = errors.New("unknown action")

type WinRegTask struct {
	ActionType WinRegActionType
	TypeName   string
	Path       string

	Name    string `taco:"name"`
	RegPath string `taco:"reg_path"`
	Val     string `taco:"value"`
	ValType string `taco:"type"`

	Require []string `taco:"require"`
	Creates []string `taco:"creates"`
	OnlyIf  []string `taco:"onlyif"`
	Unless  []string `taco:"unless"`

	Shell string `taco:"shell"`

	Updated bool
}

func (wrt *WinRegTask) GetTypeName() string {
	return wrt.TypeName
}

func (wrt *WinRegTask) GetRequirements() []string {
	return wrt.Require
}

func (wrt *WinRegTask) Validate(goos string) error {
	errs := &utils.Errors{}

	if wrt.ActionType == 0 {
		errs.Add(fmt.Errorf("unknown win_reg task type: %s", wrt.TypeName))
		return errs.ToError()
	}

	err := tasks.ValidateRequired(wrt.RegPath, wrt.Path+"."+tasks.RegPathField)
	if err != nil {
		errs.Add(err)
		return errs.ToError()
	}

	err = winreg.HasValidRootKey(wrt.RegPath)
	if err != nil {
		errs.Add(err)
	}

	if wrt.ActionType == ActionWinRegPresent || wrt.ActionType == ActionWinRegAbsent {
		err = tasks.ValidateRequired(wrt.Name, wrt.Path+"."+tasks.NameField)
		if err != nil {
			errs.Add(err)
		}
	}

	if wrt.ActionType == ActionWinRegPresent {
		err = tasks.ValidateRequired(wrt.Val, wrt.Path+"."+tasks.ValField)
		if err != nil {
			errs.Add(err)
		}

		err = tasks.ValidateRequired(wrt.ValType, wrt.Path+"."+tasks.ValTypeField)
		if err != nil {
			errs.Add(err)
		}
	}

	return errs.ToError()
}

func (wrt *WinRegTask) GetPath() string {
	return wrt.Path
}

func (wrt *WinRegTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", wrt.TypeName, wrt.GetPath())
}

func (wrt *WinRegTask) GetOnlyIfCmds() []string {
	return wrt.OnlyIf
}

func (wrt *WinRegTask) GetUnlessCmds() []string {
	return wrt.Unless
}

func (wrt *WinRegTask) GetCreatesFilesList() []string {
	return wrt.Creates
}

type WinRegTaskExecutor struct {
	Runner    tacoexec.Runner
	FsManager *utils.FsManager
}

func (wrte *WinRegTaskExecutor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	execRes := executionresult.ExecutionResult{
		Name:    task.GetTypeName(),
		Comment: "registry not updated",
		Changes: make(map[string]string),
	}

	if runtime.GOOS != "windows" {
		execRes.Err = errors.New("win_reg tasks only supported on Microsoft Windows")
		return execRes
	}

	logrus.Debugf("will trigger '%s' task", task.GetPath())

	wrt, ok := task.(*WinRegTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to WinRegTask", task)
		return execRes
	}

	execRes.Name = wrt.GetTypeName()

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		Path:         wrt.Path,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		Shell:        wrt.Shell,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := tasks.CheckConditionals(execCtx, wrte.FsManager, wrte.Runner, wrt)
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

	err = wrte.ExecuteTask(ctx, wrt, &execRes)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", execRes.Name, execRes.Duration)

	return execRes
}

func (wrte *WinRegTaskExecutor) ExecuteTask(ctx context.Context, t *WinRegTask, res *executionresult.ExecutionResult) (err error) {
	var updated bool
	var desc string

	switch t.ActionType {
	case ActionWinRegPresent:
		updated, desc, err = winreg.SetValue(t.RegPath, t.Name, t.Val, winreg.REG_SZ)
		if err != nil {
			res.Err = err
			return err
		}
		t.Updated = updated
	case ActionWinRegAbsent:
		updated, desc, err = winreg.RemoveValue(t.RegPath, t.Name)
		if err != nil {
			res.Err = err
			return err
		}
		t.Updated = updated
	case ActionWinRegAbsentKey:
		updated, desc, err = winreg.RemoveKey(t.RegPath)
		if err != nil {
			res.Err = err
			return err
		}
		t.Updated = updated
	default:
		res.Err = ErrUnknownWinRegAction
		return err
	}

	if t.Updated {
		res.Comment = "registry updated"
		res.Changes["registry"] = desc
	}

	return nil
}
