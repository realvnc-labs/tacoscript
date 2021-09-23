package tasks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"
	"gopkg.in/yaml.v2"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/sirupsen/logrus"
)

type CmdRunTask struct {
	TypeName string
	Path     string
	NamedTask
	WorkingDir            string
	User                  string
	Shell                 string
	Envs                  conv.KeyValues
	MissingFilesCondition []string
	Require               []string
	OnlyIf                []string
	Unless                []string

	// aborts task execution if one task fails
	AbortOnError bool

	// aborts execution of a multi command task if a command fails
	StopOnError bool
}

type CmdRunTaskBuilder struct {
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, ctx interface{}) (Task, error) {
	t := &CmdRunTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := utils.Errors{}

	for _, item := range ctx.([]interface{}) {
		row := item.(yaml.MapSlice)[0]
		key := row.Key.(string)
		val := row.Value

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
		case AbortOnErrorField:
			t.AbortOnError = conv.ConvertToBool(val)
		case StopOnErrorField:
			t.StopOnError = conv.ConvertToBool(val)
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
	return conv.ConvertSourceToJSONStrIfPossible(crt)
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

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec2.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   cmdRunTask.WorkingDir,
		User:         cmdRunTask.User,
		Path:         cmdRunTask.Path,
		Envs:         cmdRunTask.Envs,
		Cmds:         cmdRunTask.GetNames(),
		Shell:        cmdRunTask.Shell,
		StopOnError:  cmdRunTask.StopOnError,
	}

	shouldNotBeExecutedReason, err := crte.shouldBeExecuted(execCtx, cmdRunTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if shouldNotBeExecutedReason != "" {
		execRes.IsSkipped = true
		execRes.SkipReason = shouldNotBeExecutedReason
		return execRes
	}

	start := time.Now()

	// XXXX respect stop_on_error

	err = crte.Runner.Run(execCtx)

	execRes.Duration = time.Since(start)
	if err != nil {
		execRes.Err = err
	}
	logrus.Debugf("execution of %s has finished, took: %v", cmdRunTask.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()
	execRes.Pids = execCtx.Pids

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

	logrus.Debugf("any unless condition didn't fail for task '%s'", cmdRunTask.Path)
	return false, nil
}

func (crte *CmdRunTaskExecutor) shouldBeExecuted(ctx *exec2.Context, cmdRunTask *CmdRunTask) (skipExecutionReason string, err error) {
	isExists, filename, err := crte.checkMissingFileCondition(cmdRunTask)
	if err != nil {
		return "", err
	}

	if isExists {
		skipExecutionReason = fmt.Sprintf("file %s exists", filename)
		logrus.Debugf(skipExecutionReason+", will skip the execution of %s", cmdRunTask.Path)
		return skipExecutionReason, nil
	}

	isSuccess, err := crte.checkOnlyIfs(ctx, cmdRunTask)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return onlyIfConditionFailedReason, nil
	}

	isExpectationSuccess, err := crte.checkUnless(ctx, cmdRunTask)
	if err != nil {
		return "", err
	}

	if !isExpectationSuccess {
		skipExecutionReason = "unless condition is true"
		logrus.Debugf(skipExecutionReason+", will skip %s", cmdRunTask.Path)
		return skipExecutionReason, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", cmdRunTask.Path)
	return "", nil
}

func (crte *CmdRunTaskExecutor) checkMissingFileCondition(cmdRunTask *CmdRunTask) (isExists bool, filename string, err error) {
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
			return true, missingFileCondition, nil
		}
		logrus.Debugf("file '%s' doesn't exist", missingFileCondition)
	}

	return
}
