package tasks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/sirupsen/logrus"
)

type CmdRunTask struct {
	Names                 []string
	TypeName              string
	Path                  string
	Name                  string
	WorkingDir            string
	User                  string
	Shell                 string
	Envs                  conv.KeyValues
	MissingFilesCondition []string
	Require               []string
	OnlyIf                []string
	Unless                []string
}

type CmdRunTaskBuilder struct {
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, ctx []map[string]interface{}) (Task, error) {
	t := &CmdRunTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := utils.Errors{}
	for _, contextItem := range ctx {
		for key, val := range contextItem {
			var err error
			switch key {
			case NameField:
				t.Name = fmt.Sprint(val)
			case CwdField:
				t.WorkingDir = fmt.Sprint(val)
			case UserField:
				t.User = fmt.Sprint(val)
			case ShellField:
				t.Shell = fmt.Sprint(val)
			case EnvField:
				var envs conv.KeyValues
				envs, err = conv.ConvertToKeyValues(val, path)
				errs.Add(err)
				t.Envs = envs
			case CreatesField:
				t.MissingFilesCondition, err = parseCreatesField(val, path)
				errs.Add(err)
			case NamesField:
				var names []string
				names, err = conv.ConvertToValues(val, path)
				errs.Add(err)
				t.Names = names
			case RequireField:
				t.Require, err = parseRequireField(val, path)
				errs.Add(err)
			case OnlyIf:
				t.OnlyIf, err = parseOnlyIfField(val, path)
				errs.Add(err)
			case Unless:
				t.Unless, err = parseUnlessField(val, path)
				errs.Add(err)
			}
		}
	}

	return t, errs.ToError()
}

func (crt *CmdRunTask) GetName() string {
	return crt.TypeName
}

func (crt *CmdRunTask) GetRequirements() []string {
	return crt.Require
}

func (crt *CmdRunTask) Validate() error {
	errs := &utils.Errors{}
	err1 := ValidateRequired(crt.Name, crt.Path+"."+NameField)
	err2 := ValidateRequiredMany(crt.Names, crt.Path+"."+NamesField)

	if err1 != nil && err2 != nil {
		errs.Add(err1)
		errs.Add(err2)
		return errs.ToError()
	}

	return nil
}

func (crt *CmdRunTask) GetPath() string {
	return crt.Path
}

func (crt *CmdRunTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", crt.TypeName, crt.GetPath())
}

type CmdRunTaskExecutor struct {
	Runner    exec2.Runner
	FsManager FsManager
}

func (crte *CmdRunTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	execRes := ExecutionResult{}
	cmdRunTask, ok := task.(*CmdRunTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to CmdRunTask", task)
		return execRes
	}

	rawCmds := []string{cmdRunTask.Name}
	rawCmds = append(rawCmds, cmdRunTask.Names...)

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec2.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   cmdRunTask.WorkingDir,
		User:         cmdRunTask.User,
		Path:         cmdRunTask.Path,
		Envs:         cmdRunTask.Envs,
		Cmds:         rawCmds,
		Shell:        cmdRunTask.Shell,
	}

	shouldBeExecuted, err := crte.shouldBeExecuted(execCtx, cmdRunTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !shouldBeExecuted {
		execRes.IsSkipped = true
		return execRes
	}

	start := time.Now()

	err = crte.Runner.Run(execCtx)

	execRes.Duration = time.Since(start)
	if err != nil {
		execRes.Err = err
	}
	logrus.Debugf("execution of %s has finished, took: %v", cmdRunTask.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()

	return execRes
}

func (crte *CmdRunTaskExecutor) checkOnlyIfs(ctx *exec2.Context, cmdRunTask *CmdRunTask) (isSuccess bool, err error) {
	if len(cmdRunTask.OnlyIf) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = cmdRunTask.OnlyIf
	err = crte.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Debugf("will skip %s since onlyif condition has failed: %v", cmdRunTask.Path, runErr)
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (crte *CmdRunTaskExecutor) checkUnless(ctx *exec2.Context, cmdRunTask *CmdRunTask) (isExpectationSuccess bool, err error) {
	if len(cmdRunTask.Unless) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = cmdRunTask.Unless

	err = crte.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Infof("will continue cmd since at least one unless condition has failed: %v", runErr)
			return true, nil
		}

		return false, err
	}

	logrus.Infof("any unless condition didn't fail for task '%s'", cmdRunTask.Path)
	return false, nil
}

func (crte *CmdRunTaskExecutor) shouldBeExecuted(ctx *exec2.Context, cmdRunTask *CmdRunTask) (shouldBeExecuted bool, err error) {
	isExists, err := crte.checkMissingFileCondition(cmdRunTask)
	if err != nil {
		return false, err
	}

	if isExists {
		logrus.Debugf("some files exist, will skip the execution of %s", cmdRunTask.Path)
		return false, nil
	}

	isSuccess, err := crte.checkOnlyIfs(ctx, cmdRunTask)
	if err != nil {
		return false, err
	}

	if !isSuccess {
		return false, nil
	}

	isExpectationSuccess, err := crte.checkUnless(ctx, cmdRunTask)
	if err != nil {
		return false, err
	}

	if !isExpectationSuccess {
		logrus.Debugf("check of unless section was false, will skip %s", cmdRunTask.Path)
		return false, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", cmdRunTask.Path)
	return true, nil
}

func (crte *CmdRunTaskExecutor) checkMissingFileCondition(cmdRunTask *CmdRunTask) (isExists bool, err error) {
	if len(cmdRunTask.MissingFilesCondition) == 0 {
		return
	}

	for _, missingFileCondition := range cmdRunTask.MissingFilesCondition {
		if missingFileCondition == "" {
			continue
		}
		isExists, err = crte.FsManager.FileExists(missingFileCondition)
		if err != nil {
			err = fmt.Errorf("failed to check if file '%s' exists: %w", missingFileCondition, err)
			return
		}

		if isExists {
			logrus.Debugf("file '%s' exists", missingFileCondition)
			return
		}
		logrus.Debugf("file '%s' doesn't exist", missingFileCondition)
	}

	return
}
