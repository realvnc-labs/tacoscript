package tasks

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
		t.Mode, err = conv.ConvertToFileMode(val)
		if err != nil {
			return err
		}
	case EncodingField:
		t.Encoding = fmt.Sprint(val)
	case ContentsField:
		t.Contents = fmt.Sprint(val)
	}

	return nil
}

type FileManagedTask struct {
	MakeDirs     bool
	Replace      bool
	SkipVerify   bool
	SkipTLSCheck bool
	Mode         os.FileMode
	TypeName     string
	Path         string
	Name         string
	SourceHash   string
	Contents     string
	User         string
	Group        string
	Encoding     string
	Source       utils.Location
	Creates      []string
	OnlyIf       []string
	Require      []string
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

	err = fmte.copySourceToTarget(ctx, fileManagedTask)
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
		var hashEquals bool
		hashEquals, _, err = utils.HashEquals(fileManagedTask.SourceHash, fileManagedTask.Name)
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

func (fmte *FileManagedTaskExecutor) copySourceToTarget(ctx context.Context, fileManagedTask *FileManagedTask) error {
	if fileManagedTask.Source.RawLocation == "" {
		logrus.Debug("source location is empty will ignore it")
		return nil
	}

	logrus.Debugf("will copy source location '%s' to target location '%s'", fileManagedTask.Source.RawLocation, fileManagedTask.Name)

	source := fileManagedTask.Source
	if !source.IsURL {
		logrus.Debug("source location is a local file path")

		hashEquals, expectedHashStr, err := utils.HashEquals(fileManagedTask.SourceHash, source.LocalPath)
		if err != nil {
			return err
		}
		if !hashEquals {
			return fmt.Errorf(
				"checksum '%s' didn't match with checksum '%s' of the local source '%s'",
				fileManagedTask.SourceHash,
				expectedHashStr,
				source.LocalPath,
			)
		}

		return utils.CopyLocalFile(source.LocalPath, fileManagedTask.Name)
	}

	tempTargetPath := fileManagedTask.Name + "_temp"
	defer func() {
		err := os.Remove(tempTargetPath)
		if err != nil {
			logrus.Warn(err)
		}
	}()

	logrus.Debug("source location is a remote file path")

	var err error
	switch fileManagedTask.Source.URL.Scheme {
	case "http":
		err = utils.DownloadHTTPFile(ctx, fileManagedTask.Source.URL, tempTargetPath)
	case "https":
		err = utils.DownloadHTTPSFile(ctx, fileManagedTask.SkipTLSCheck, fileManagedTask.Source.URL, tempTargetPath)
	case "ftp":
		err = utils.DownloadFtpFile(ctx, fileManagedTask.Source.URL, tempTargetPath)
	default:
		err = fmt.Errorf(
			"unknown or unsupported protocol '%s' to download data from '%s'",
			fileManagedTask.Source.URL.Scheme,
			fileManagedTask.Source.URL,
		)
	}

	if err != nil {
		return err
	}

	logrus.Debugf("copied remove source '%s' to a temp location '%s', will check the hash", source.RawLocation, tempTargetPath)

	hashEquals, expectedHashStr, err := utils.HashEquals(fileManagedTask.SourceHash, tempTargetPath)
	if err != nil {
		return err
	}
	if !hashEquals {
		return fmt.Errorf(
			"checksum '%s' didn't match with checksum '%s' of the remote source '%s'",
			fileManagedTask.SourceHash,
			expectedHashStr,
			source.RawLocation,
		)
	}

	logrus.Debug("checksum file at temp location matched with the expected one")
	logrus.Debugf("will move file from temp location '%s' to the expected location '%s'", tempTargetPath, fileManagedTask.Name)

	err = utils.MoveFile(tempTargetPath, fileManagedTask.Name)
	if err != nil {
		return err
	}

	logrus.Debugf("copied field from temp location '%s' to the expected location '%s'", tempTargetPath, fileManagedTask.Name)

	return nil
}
