package tasks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
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
	}

	errs := &utils.Errors{}
	for _, contextItem := range ctx {
		for key, val := range contextItem {
			err := fmtb.processContextItem(t, key, path, val)
			errs.Add(err)
		}
	}

	return t, errs.ToError()
}

func (fmtb FileManagedTaskBuilder) processContextItem(t *FileManagedTask, key, path string, val interface{}) error {
	var err error
	switch key {
	case NameField:
		t.Name = fmt.Sprint(val)
	case UserField:
		t.User = fmt.Sprint(val)
	case CreatesField:
		t.Creates, err = parseCreatesField(val, path)
		return err
	case RequireField:
		t.Require, err = parseRequireField(val, path)
		return err
	case OnlyIf:
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	case SkipVerifyField:
		t.SkipVerify = conv.ConvertToBool(val)
	case SourceField:
		t.Source = utils.ParseLocation(fmt.Sprint(val))
	case SourceHashField:
		t.SourceHash = fmt.Sprint(val)
	case MakeDirsField:
		t.MakeDirs = conv.ConvertToBool(val)
	case GroupField:
		t.Group = fmt.Sprint(val)
	case ModeField:
		fileUint, ok := val.(int)
		if ok {
			t.Mode = os.FileMode(fileUint)
			return nil
		}

		valStr := fmt.Sprint(val)
		i64, err := strconv.ParseInt(valStr, 8, 32)
		if err != nil {
			return fmt.Errorf(`invalid file mode value '%s' at path 'invalid_filemode_path.%s'`, valStr, ModeField)
		}
		t.Mode = os.FileMode(i64)
	case EncodingField:
		t.Encoding = fmt.Sprint(val)
	case ContentsField:
		t.Contents = fmt.Sprint(val)
	}

	return nil
}

type FileManagedTask struct {
	TypeName     string
	Path         string
	Name         string
	Source       utils.Location
	SourceHash   string
	MakeDirs     bool
	Replace      bool
	SkipVerify   bool
	Creates      []string
	Contents     string
	User         string
	Group        string
	Encoding     string
	Mode         os.FileMode
	OnlyIf       []string
	Require      []string
	SkipTlsCheck bool
}

func (crt *FileManagedTask) GetName() string {
	return crt.TypeName
}

func (crt *FileManagedTask) GetRequirements() []string {
	return crt.Require
}

func (crt *FileManagedTask) Validate() error {
	errs := &utils.Errors{}

	err1 := ValidateRequired(crt.Name, crt.Path+"."+NameField)
	errs.Add(err1)

	if crt.Source.IsURL && crt.SourceHash == "" {
		errs.Add(
			fmt.Errorf(
				`empty '%s' field at path '%s.%s' for remote url source '%s'`,
				SourceHashField,
				crt.Path,
				SourceHashField,
				crt.Source.RawLocation,
			),
		)
	}

	return errs.ToError()
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

	err = fmte.copySourceToTarget(fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	execRes.Duration = time.Since(start)

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

	if fileManagedTask.SourceHash != "" {
		hashEquals, err := utils.HashEquals(fileManagedTask.SourceHash, fileManagedTask.Name)
		if err != nil {
			return false, err
		}
		if hashEquals {
			logrus.Debugf("hash '%s' matches the hash sum of file at '%s', will not update it", fileManagedTask.SourceHash, fileManagedTask.Name)
			return false, nil
		}
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
		isExists, err = fmte.FsManager.FileExists(missingFileCondition)
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

func (fmte *FileManagedTaskExecutor) copySourceToTarget(fileManagedTask *FileManagedTask) error {
	if fileManagedTask.Source.RawLocation == "" {
		logrus.Debug("source location is empty will ignore it")
		return nil
	}

	source := fileManagedTask.Source
	if !source.IsURL {
		return utils.CopyLocalFile(source.LocalPath, fileManagedTask.Name)
	}

	switch fileManagedTask.Source.Url.Scheme {
	case "http":
		return utils.DownloadHttpFile(fileManagedTask.Source.Url, fileManagedTask.Name)
	case "https":
		return utils.DownloadHttpsFile(fileManagedTask.SkipTlsCheck, fileManagedTask.Source.Url, fileManagedTask.Name)
	case "ftp":
		return utils.DownloadFtpFile(fileManagedTask.Source.Url, fileManagedTask.Name)
	default:
		return fmt.Errorf("unknown or unsupported protocol '%s' to download data from '%s'", fileManagedTask.Source.Url.Scheme, fileManagedTask.Source.Url)
	}
}
