package tasks

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

const DefaultFileMode = 0744

type FileManagedTaskBuilder struct {
}

type contextProc func(t *FileManagedTask, path string, val interface{}) error

var contextProcMap = map[string]contextProc{
	NameField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Name = fmt.Sprint(val)
		return nil
	},
	UserField: func(t *FileManagedTask, path string, val interface{}) error {
		t.User = fmt.Sprint(val)
		return nil
	},
	CreatesField: func(t *FileManagedTask, path string, val interface{}) error {
		var err error
		t.Creates, err = parseCreatesField(val, path)
		return err
	},
	RequireField: func(t *FileManagedTask, path string, val interface{}) error {
		var err error
		t.Require, err = parseRequireField(val, path)
		return err
	},
	OnlyIf: func(t *FileManagedTask, path string, val interface{}) error {
		var err error
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	},
	SkipVerifyField: func(t *FileManagedTask, path string, val interface{}) error {
		t.SkipVerify = conv.ConvertToBool(val)
		return nil
	},
	SourceField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Source = utils.ParseLocation(fmt.Sprint(val))
		return nil
	},
	SourceHashField: func(t *FileManagedTask, path string, val interface{}) error {
		t.SourceHash = fmt.Sprint(val)
		return nil
	},
	MakeDirsField: func(t *FileManagedTask, path string, val interface{}) error {
		t.MakeDirs = conv.ConvertToBool(val)
		return nil
	},
	GroupField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Group = fmt.Sprint(val)
		return nil
	},
	ModeField: func(t *FileManagedTask, path string, val interface{}) error {
		var err error
		t.Mode, err = conv.ConvertToFileMode(val)
		return err
	},
	EncodingField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Encoding = fmt.Sprint(val)
		return nil
	},
	ContentsField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Contents = parseContentsField(val)
		return nil
	},
	ReplaceField: func(t *FileManagedTask, path string, val interface{}) error {
		t.Replace = conv.ConvertToBool(val)
		return nil
	},
}

func (fmtb FileManagedTaskBuilder) Build(typeName, path string, ctx []map[string]interface{}) (Task, error) {
	t := &FileManagedTask{
		TypeName: typeName,
		Path:     path,
		Replace:  true,
	}

	errs := &utils.Errors{}
	for _, contextItem := range ctx {
		for key, val := range contextItem {
			f, ok := contextProcMap[key]
			if !ok {
				continue
			}
			errs.Add(f(t, path, val))
		}
	}

	return t, errs.ToError()
}

func parseContentsField(val interface{}) sql.NullString {
	isValid := false
	if val != nil {
		isValid = true
	}
	return sql.NullString{
		String: fmt.Sprint(val),
		Valid:  isValid,
	}
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
	Contents     sql.NullString
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

	if crt.Source.IsURL && crt.SourceHash == "" && !crt.SkipVerify {
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

	if crt.Source.RawLocation == "" && !crt.Contents.Valid {
		errs.Add(fmt.Errorf(
			`either content or source should be provided for the task at path '%s'`,
			crt.Path,
		))
	}

	return errs.ToError()
}

func (crt *FileManagedTask) GetPath() string {
	return crt.Path
}

func (crt *FileManagedTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", crt.TypeName, crt.GetPath())
}

type HashManager interface {
	HashEquals(hashStr, filePath string) (hashEquals bool, actualCache string, err error)
	HashSum(hashAlgoName, filePath string) (hashSum string, err error)
}

type FileManagedTaskExecutor struct {
	FsManager   FsManager
	HashManager HashManager
	Runner      exec2.Runner
}

func (fmte *FileManagedTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
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
	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	shouldBeExecuted, err := fmte.shouldBeExecuted(execCtx, fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !shouldBeExecuted {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		return execRes
	}

	start := time.Now()

	fileShouldBeReplaced, err := fmte.fileShouldBeReplaced(fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if fileShouldBeReplaced {
		err = fmte.createDirPathIfNeeded(fileManagedTask)
		if err != nil {
			execRes.Err = err
			return execRes
		}

		err = fmte.copySourceToTarget(ctx, fileManagedTask)
		if err != nil {
			execRes.Err = err
			return execRes
		}

		err = fmte.copyContentToTarget(fileManagedTask)
		if err != nil {
			execRes.Err = err
			return execRes
		}
	}

	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", task.GetPath(), execRes.Duration)
	return execRes
}

func (fmte *FileManagedTaskExecutor) fileShouldBeReplaced(fileManagedTask *FileManagedTask) (bool, error) {
	if fileManagedTask.Replace {
		return true, nil
	}

	fileExists, err := fmte.FsManager.FileExists(fileManagedTask.Name)
	if err != nil {
		return true, err
	}

	if fileExists {
		logrus.Debugf("since file '%s' exists and '%s' field is set to false, file won't be changed", fileManagedTask.Name, ReplaceField)
		return false, nil
	}

	return true, nil
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
		hashEquals, _, err = fmte.HashManager.HashEquals(fileManagedTask.SourceHash, fileManagedTask.Name)
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

	shouldSkip, err := fmte.shouldSkipForContentExpectation(fileManagedTask)
	if err != nil {
		return false, err
	}
	if shouldSkip {
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
	source := fileManagedTask.Source
	if source.RawLocation == "" {
		logrus.Debug("source location is empty will ignore it")
		return nil
	}

	logrus.Debugf("will copy source location '%s' to target location '%s'", fileManagedTask.Source.RawLocation, fileManagedTask.Name)

	if !source.IsURL {
		return fmte.handleLocalSource(fileManagedTask, source.LocalPath)
	}

	return fmte.handleRemoteSource(ctx, fileManagedTask)
}

func (fmte *FileManagedTaskExecutor) handleRemoteSource(ctx context.Context, fileManagedTask *FileManagedTask) error {
	tempTargetPath := fileManagedTask.Name + "_temp"

	logrus.Debug("source location is a remote url")

	defer func(f string) {
		fileExists, err := fmte.FsManager.FileExists(f)
		if !fileExists || err != nil {
			return
		}

		err = fmte.FsManager.Remove(f)
		if err != nil {
			logrus.Errorf("failed to delete '%s': %v", f, err)
		}
	}(tempTargetPath)
	err := fmte.FsManager.DownloadFile(ctx, tempTargetPath, fileManagedTask.Source.URL, fileManagedTask.SkipTLSCheck)
	if err != nil {
		return err
	}
	logrus.Debugf(
		"copied remove source '%s' to a temp location '%s', will check the hash",
		fileManagedTask.Source.RawLocation,
		tempTargetPath,
	)

	shouldBeCopied, err := fmte.checkIfLocalFileShouldBeCopied(fileManagedTask, tempTargetPath)
	if err != nil {
		return err
	}
	if !shouldBeCopied {
		return nil
	}

	err = fmte.FsManager.MoveFile(tempTargetPath, fileManagedTask.Name)
	if err != nil {
		return err
	}

	logrus.Debugf(
		"copied field from temp location '%s' to the expected location '%s'",
		tempTargetPath,
		fileManagedTask.Name,
	)

	return nil
}

func (fmte *FileManagedTaskExecutor) handleLocalSource(fileManagedTask *FileManagedTask, sourcePath string) error {
	logrus.Debug("source location is a local file path")
	source := fileManagedTask.Source

	shouldBeCopied, err := fmte.checkIfLocalFileShouldBeCopied(fileManagedTask, sourcePath)
	if err != nil {
		return err
	}
	if !shouldBeCopied {
		return nil
	}

	return fmte.FsManager.CopyLocalFile(source.LocalPath, fileManagedTask.Name)
}

func (fmte *FileManagedTaskExecutor) checkIfLocalFileShouldBeCopied(fileManagedTask *FileManagedTask, sourcePath string) (bool, error) {
	const defaultHashAlgoName = "sha256"

	if !fileManagedTask.SkipVerify {
		hashEquals, expectedHashStr, err := fmte.HashManager.HashEquals(fileManagedTask.SourceHash, sourcePath)
		if err != nil {
			return false, err
		}
		if !hashEquals {
			logrus.Debugf(
				"expected source hash '%s' didn't match with the source file '%s' which means source "+
					"was unexpectedly modified, will report as an error",
				fileManagedTask.SourceHash,
				sourcePath,
			)
			return false, fmt.Errorf(
				"expected hash sum '%s' didn't match with checksum '%s' of the source file '%s'",
				fileManagedTask.SourceHash,
				expectedHashStr,
				sourcePath,
			)
		}
	}

	logrus.Debug("since skip verify is set to true will ignore source hash and check if the hash sum " +
		"of the local source file matches with the hash sum of the target file")
	sourceFileHashSum, err := fmte.HashManager.HashSum(defaultHashAlgoName, sourcePath)
	if err != nil {
		return false, err
	}

	fileExists, err := fmte.FsManager.FileExists(fileManagedTask.Name)
	if err != nil {
		return false, err
	}

	if !fileExists {
		logrus.Debugf("since local target file '%s' doesn't exist, it should be created with the source file contents", fileManagedTask.Name)
		return true, nil
	}

	targetFileHashSum, err := fmte.HashManager.HashSum(defaultHashAlgoName, fileManagedTask.Name)
	if err != nil {
		return false, err
	}

	if sourceFileHashSum != targetFileHashSum {
		logrus.Debugf(
			"target file '%s' hash sum[%s] '%s' didn't match with the source file '%s' hash sum '%s', so contents of source should be copied",
			fileManagedTask.Name,
			defaultHashAlgoName,
			targetFileHashSum,
			sourcePath,
			sourceFileHashSum,
		)
		return true, nil
	}

	logrus.Debugf(
		"target file '%s' hash sum[%s] '%s' matches with the source file '%s' hash sum, so target should not be changed",
		fileManagedTask.Name,
		defaultHashAlgoName,
		targetFileHashSum,
		sourceFileHashSum,
	)

	return false, nil
}

func (fmte *FileManagedTaskExecutor) copyContentToTarget(fileManagedTask *FileManagedTask) error {
	if !fileManagedTask.Contents.Valid {
		logrus.Debug("contents field is empty, will not copy data")
		return nil
	}

	mode := os.FileMode(DefaultFileMode)
	logrus.Debugf("will write contents to target file '%s'", fileManagedTask.Name)
	err := fmte.FsManager.WriteFile(fileManagedTask.Name, fileManagedTask.Contents.String, mode)

	if err == nil {
		logrus.Debugf("written contents to '%s'", fileManagedTask.Name)
	}

	return err
}

func (fmte *FileManagedTaskExecutor) shouldSkipForContentExpectation(fileManagedTask *FileManagedTask) (bool, error) {
	if !fileManagedTask.Contents.Valid {
		logrus.Debug("contents section is missing, won't check the content")
		return false, nil
	}

	logrus.Debugf("will compare contents of file '%s' with the provided contents", fileManagedTask.Name)
	actualContents := ""

	fileExists, err := fmte.FsManager.FileExists(fileManagedTask.Name)
	if err != nil {
		return false, err
	}

	if fileExists {
		actualContents, err = fmte.FsManager.ReadFile(fileManagedTask.Name)
		if err != nil {
			return false, err
		}
	}

	contentDiff := utils.Diff(fileManagedTask.Contents.String, actualContents)
	if contentDiff == "" {
		logrus.Debugf("file '%s' matched with the expected contents, will skip the execution", fileManagedTask.Name)
		return true, nil
	}

	logrus.WithFields(
		logrus.Fields{
			"multiline": contentDiff,
		}).Infof(`file '%s' differs from the expected content field, will copy diff to file`, fileManagedTask.Name)

	return false, nil
}

func (fmte *FileManagedTaskExecutor) createDirPathIfNeeded(fileManagedTask *FileManagedTask) error {
	if !fileManagedTask.MakeDirs {
		return nil
	}

	logrus.Debugf("will create dirs for '%s' if needed", fileManagedTask.Name)

	var mode os.FileMode
	if fileManagedTask.Mode == 0 {
		mode = DefaultFileMode
	} else {
		mode = fileManagedTask.Mode
	}

	return fmte.FsManager.CreateDirPathIfNeeded(fileManagedTask.Name, mode)
}
