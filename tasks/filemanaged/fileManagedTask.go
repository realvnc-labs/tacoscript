package filemanaged

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"

	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/sirupsen/logrus"
)

const (
	TaskTypeFileManaged = "file.managed"

	DefaultFileMode = 0744
)

type FmTask struct {
	TypeName string
	Path     string
	Mode     os.FileMode
	Contents sql.NullString
	Source   utils.Location

	Name         string   `taco:"name"`
	MakeDirs     bool     `taco:"makedirs"`
	Replace      bool     `taco:"replace"`
	SkipVerify   bool     `taco:"skip_verify"`
	SkipTLSCheck bool     `taco:"???"`
	SourceHash   string   `taco:"source_hash"`
	User         string   `taco:"user"`
	Group        string   `taco:"group"`
	Encoding     string   `taco:"encoding"`
	Creates      []string `taco:"creates"`
	OnlyIf       []string `taco:"onlyif"`
	Unless       []string `taco:"unless"`
	Require      []string `taco:"require"`

	Shell string `taco:"shell"`

	// was managed file updated?
	Updated bool
}

func (t *FmTask) GetTypeName() string {
	return t.TypeName
}

func (t *FmTask) GetRequirements() []string {
	return t.Require
}

func (t *FmTask) Validate(goos string) error {
	errs := &utils.Errors{}

	err1 := tasks.ValidateRequired(t.Name, t.Path+"."+tasks.NameField)
	errs.Add(err1)

	if t.Source.IsURL && t.SourceHash == "" && !t.SkipVerify {
		errs.Add(
			fmt.Errorf(
				`empty '%s' field at path '%s.%s' for remote url source '%s'`,
				tasks.SourceHashField,
				t.Path,
				tasks.SourceHashField,
				t.Source.RawLocation,
			),
		)
	}

	if t.Source.RawLocation == "" && !t.Contents.Valid {
		errs.Add(fmt.Errorf(
			`either content or source should be provided for the task at path '%s'`,
			t.Path,
		))
	}

	return errs.ToError()
}

func (t *FmTask) GetPath() string {
	return t.Path
}

func (t *FmTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.TypeName, t.GetPath())
}

func (t *FmTask) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *FmTask) GetUnlessCmds() []string {
	return t.Unless
}

func (t *FmTask) GetCreatesFilesList() []string {
	return t.Creates
}

type HashManager interface {
	HashEquals(hashStr, filePath string) (hashEquals bool, actualCache string, err error)
	HashSum(hashAlgoName, filePath string) (hashSum string, err error)
}

type FmtExecutor struct {
	FsManager   tasks.FsManager
	HashManager HashManager
	Runner      tacoexec.Runner
}

func (fmte *FmtExecutor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := executionresult.ExecutionResult{
		Changes: make(map[string]string),
	}

	fileManagedTask, ok := task.(*FmTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to FmTask", task)
		return execRes
	}

	execRes.Name = fileManagedTask.Name

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		User:         fileManagedTask.User,
		Path:         fileManagedTask.Path,
		Shell:        fileManagedTask.Shell,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())

	skipReason, err := tasks.CheckConditionals(execCtx, fmte.FsManager, fmte.Runner, fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	// if core conditionals ok, then check the specific file managed conditions
	if err == nil && skipReason == "" {
		skipReason, err = fmte.checkFileManagedConditions(fileManagedTask, &execRes)
		if err != nil {
			execRes.Err = err
			return execRes
		}
	}

	if skipReason != "" {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		execRes.SkipReason = skipReason
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

		var info fs.FileInfo
		info, err = fmte.FsManager.Stat(fileManagedTask.Name)
		if err != nil {
			execRes.Err = err
			return execRes
		}

		fileManagedTask.Updated = true
		execRes.Comment = "File updated"
		execRes.Changes["length"] = fmt.Sprintf("%d bytes written", info.Size())
	}

	err = fmte.applyFileAttributesToTarget(fileManagedTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", task.GetPath(), execRes.Duration)
	return execRes
}

func (fmte *FmtExecutor) fileShouldBeReplaced(fileManagedTask *FmTask) (bool, error) {
	if fileManagedTask.Replace {
		return true, nil
	}

	fileExists, err := fmte.FsManager.FileExists(fileManagedTask.Name)
	if err != nil {
		return true, err
	}

	if fileExists {
		logrus.Debugf("since file '%s' exists and '%s' field is set to false, file won't be changed", fileManagedTask.Name, tasks.ReplaceField)
		return false, nil
	}

	return true, nil
}

func (fmte *FmtExecutor) checkFileManagedConditions(
	fileManagedTask *FmTask,
	execRes *executionresult.ExecutionResult,
) (skipReason string, err error) {
	if fileManagedTask.SourceHash != "" {
		var hashEquals bool
		hashEquals, _, err = fmte.HashManager.HashEquals(fileManagedTask.SourceHash, fileManagedTask.Name)
		if err != nil {
			return "", err
		}
		if hashEquals {
			skipReason = fmt.Sprintf(
				"hash '%s' matches the hash sum of file at '%s', will not update it",
				fileManagedTask.SourceHash,
				fileManagedTask.Name,
			)
			logrus.Debug(skipReason)
			return skipReason, nil
		}
	}

	skipReasonForContents, err := fmte.shouldSkipForContentExpectation(fileManagedTask, execRes)
	if err != nil {
		return "", err
	}
	if skipReasonForContents != "" {
		return skipReasonForContents, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", fileManagedTask.String())
	return "", nil
}

func (fmte *FmtExecutor) copySourceToTarget(ctx context.Context, fileManagedTask *FmTask) error {
	source := fileManagedTask.Source
	if source.RawLocation == "" {
		logrus.Debug("source location is empty will ignore it")
		return nil
	}

	if !source.IsURL {
		return fmte.handleLocalSource(fileManagedTask, source.LocalPath)
	}

	return fmte.handleRemoteSource(ctx, fileManagedTask)
}

func (fmte *FmtExecutor) handleRemoteSource(ctx context.Context, fileManagedTask *FmTask) error {
	tempTargetPath := fileManagedTask.Name + "_temp"

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
		"copied remove source '%s' to a temp location '%s'",
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
		"moved file from a temp location '%s' to the target location '%s'",
		tempTargetPath,
		fileManagedTask.Name,
	)

	return nil
}

func (fmte *FmtExecutor) handleLocalSource(fileManagedTask *FmTask, sourcePath string) error {
	logrus.Debug("source location is a local file path")
	source := fileManagedTask.Source

	shouldBeCopied, err := fmte.checkIfLocalFileShouldBeCopied(fileManagedTask, sourcePath)
	if err != nil {
		return err
	}
	if !shouldBeCopied {
		return nil
	}

	mode := os.FileMode(DefaultFileMode)
	if fileManagedTask.Mode > 0 {
		mode = fileManagedTask.Mode
	}

	return fmte.FsManager.CopyLocalFile(source.LocalPath, fileManagedTask.Name, mode)
}

func (fmte *FmtExecutor) checkIfLocalFileShouldBeCopied(fileManagedTask *FmTask, sourcePath string) (bool, error) {
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
		return true, nil
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

func (fmte *FmtExecutor) copyContentToTarget(fileManagedTask *FmTask) error {
	if !fileManagedTask.Contents.Valid {
		logrus.Debug("contents field is empty, will not manage content")
		return nil
	}

	mode := os.FileMode(DefaultFileMode)
	if fileManagedTask.Mode > 0 {
		mode = fileManagedTask.Mode
	}

	logrus.Debugf("will write contents to target file '%s'", fileManagedTask.Name)

	var err error
	if fileManagedTask.Encoding != "" {
		logrus.Debugf("will encode file contents to '%s'", fileManagedTask.Encoding)
		err = utils.WriteEncodedFile(fileManagedTask.Encoding, fileManagedTask.Contents.String, fileManagedTask.Name, mode)
	} else {
		err = fmte.FsManager.WriteFile(fileManagedTask.Name, fileManagedTask.Contents.String, mode)
	}

	if err == nil {
		logrus.Debugf("written contents to '%s'", fileManagedTask.Name)
	}

	return err
}

func (fmte *FmtExecutor) shouldSkipForContentExpectation(
	fileManagedTask *FmTask,
	execRes *executionresult.ExecutionResult,
) (skipReason string, err error) {
	if !fileManagedTask.Contents.Valid {
		logrus.Debug("contents section is missing, won't check the content")
		return "", nil
	}

	logrus.Debugf("will compare contents of file '%s' with the provided contents", fileManagedTask.Name)
	actualContents := ""

	fileExists, err := fmte.FsManager.FileExists(fileManagedTask.Name)
	if err != nil {
		return "", err
	}

	if fileExists {
		if fileManagedTask.Encoding != "" {
			actualContents, err = fmte.FsManager.ReadEncodedFile(fileManagedTask.Encoding, fileManagedTask.Name)
		} else {
			actualContents, err = fmte.FsManager.ReadFile(fileManagedTask.Name)
		}

		if err != nil {
			return "", err
		}
	}

	contentDiff := utils.Diff(fileManagedTask.Contents.String, actualContents)
	if contentDiff == "" {
		skipReason = fmt.Sprintf("file '%s' matched with the expected contents, will skip the execution", fileManagedTask.Name)
		logrus.Debug(skipReason)
		return skipReason, nil
	}

	logrus.WithFields(
		logrus.Fields{
			"multiline": contentDiff,
		}).Debugf(`file '%s' differs from the expected content field, will copy diff to file`, fileManagedTask.Name)

	execRes.Changes["diff"] = contentDiff
	execRes.Changes["size_diff"] = fmt.Sprintf("%d bytes", len(fileManagedTask.Contents.String)-len(actualContents))
	return "", nil
}

func (fmte *FmtExecutor) createDirPathIfNeeded(fileManagedTask *FmTask) error {
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

func (fmte *FmtExecutor) applyFileAttributesToTarget(fileManagedTask *FmTask) error {
	logrus.Debugf("will change file attributes '%s'", fileManagedTask.Name)

	info, err := fmte.FsManager.Stat(fileManagedTask.Name)
	if err != nil {
		return err
	}

	if fileManagedTask.Mode > 0 && fileManagedTask.Mode != info.Mode() {
		err = fmte.FsManager.Chmod(fileManagedTask.Name, fileManagedTask.Mode)
		if err != nil {
			return err
		}
		logrus.Debugf("changed mode of '%s' to '%v'", fileManagedTask.Name, fileManagedTask.Mode)
	}

	if fileManagedTask.User != "" || fileManagedTask.Group != "" {
		logrus.Debugf("will change user '%s' or group '%s' of file '%s'", fileManagedTask.User, fileManagedTask.Group, fileManagedTask.Name)
		err = fmte.FsManager.Chown(fileManagedTask.Name, fileManagedTask.User, fileManagedTask.Group)
		if err != nil {
			return err
		}
	}

	return nil
}
