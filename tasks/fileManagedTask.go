package tasks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

type FileManagedTaskBuilder struct {
}

func (fmtb FileManagedTaskBuilder) Build(typeName, path string, ctx []map[string]interface{}) (Task, error) {
	t := &FileManagedTask{
		TypeName: typeName,
		Path:     path,
		Errors:   &utils.Errors{},
	}

	for _, contextItem := range ctx {
		for key, val := range contextItem {
			fmtb.processContextItem(t, key, path, val)
		}
	}

	return t, nil
}

func (fmtb FileManagedTaskBuilder) processContextItem(t *FileManagedTask, key, path string, val interface{}) {
	switch key {
	case NameField:
		t.Name = fmt.Sprint(val)
	case UserField:
		t.User = fmt.Sprint(val)
	case CreatesField:
		t.Creates = parseCreatesField(val, path, t.Errors)
	case RequireField:
		t.Require = parseRequireField(val, path, t.Errors)
	case OnlyIf:
		t.OnlyIf = parseOnlyIfField(val, path, t.Errors)
	case SkipVerifyField:
		t.SkipVerify = conv.ConvertToBool(val)
	case SourceField:
		t.Source = fmt.Sprint(val)
	case SourceHashField:
		t.SourceHash = fmt.Sprint(val)
	case MakeDirsField:
		t.MakeDirs = conv.ConvertToBool(val)
	case GroupField:
		t.Group = fmt.Sprint(val)
	case ModeField:
		t.Mode = fmt.Sprint(val)
	case EncodingField:
		t.Encoding = fmt.Sprint(val)
	case ContentsField:
		t.Contents = fmt.Sprint(val)
	}
}

type FileManagedTask struct {
	TypeName   string
	Path       string
	Name       string
	Source     string
	SourceHash string
	MakeDirs   bool
	Replace    bool
	SkipVerify bool
	Creates    []string
	Contents   string
	User       string
	Group      string
	Encoding   string
	Mode       string
	OnlyIf     []string
	Runner     exec2.Runner
	FsManager  utils.FsManager
	Require    []string
	Errors     *utils.Errors
}

func (crt *FileManagedTask) GetName() string {
	return crt.TypeName
}

func (crt *FileManagedTask) GetRequirements() []string {
	return crt.Require
}

func (crt *FileManagedTask) Validate() error {
	err1 := ValidateRequired(crt.Name, crt.Path+"."+NameField)
	crt.Errors.Add(err1)

	return nil
}

func (crt *FileManagedTask) GetPath() string {
	return crt.Path
}

func (crt *FileManagedTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", crt.TypeName, crt.GetPath())
}

type FileManagedTaskExecutor struct {
	FsManager utils.FsManager
	Runner    exec2.Runner
}

func (fmte *FileManagedTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	execRes := ExecutionResult{}

	fileManagedTask, ok := task.(*FileManagedTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to FileManagedTask", task)
		return execRes
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec2.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		User:         fileManagedTask.User,
		Path:         fileManagedTask.Path,
	}

	shouldBeExecuted, err := fmte.shouldBeExecuted(execCtx, fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !shouldBeExecuted {
		execRes.IsSkipped = true
		return execRes
	}

	start := time.Now()

	err = fmte.Runner.Run(execCtx)

	execRes.Duration = time.Since(start)
	if err != nil {
		execRes.Err = err
	}
	logrus.Debugf("execution of %s has finished, took: %v", fileManagedTask.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()

	return execRes
}

func (fmte *FileManagedTaskExecutor) checkOnlyIfs(ctx *exec2.Context, fileManagedTask *FileManagedTask) (isSuccess bool, err error) {
	if len(fileManagedTask.OnlyIf) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = fileManagedTask.OnlyIf
	err = fmte.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Debugf("will skip %s since onlyif condition has failed: %v", fileManagedTask, runErr)
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (fmte *FileManagedTaskExecutor) shouldBeExecuted(
	ctx *exec2.Context,
	fileManagedTask *FileManagedTask,
) (shouldBeExecuted bool, err error) {
	isExists, err := fmte.checkMissingFileCondition(fileManagedTask)
	if err != nil {
		return false, err
	}

	if isExists {
		logrus.Debugf("some files exist, will skip the execution of %s", fileManagedTask)
		return false, nil
	}

	isSuccess, err := fmte.checkOnlyIfs(ctx, fileManagedTask)
	if err != nil {
		return false, err
	}

	if !isSuccess {
		return false, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", fileManagedTask)
	return true, nil
}

func (fmte *FileManagedTaskExecutor) checkMissingFileCondition(fileManagedTask *FileManagedTask) (isExists bool, err error) {
	if len(fileManagedTask.Creates) == 0 {
		return
	}

	for _, missingFileCondition := range fileManagedTask.Creates {
		if missingFileCondition == "" {
			continue
		}
		isExists, err = fileManagedTask.FsManager.FileExists(missingFileCondition)
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
