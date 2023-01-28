package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/conv"
	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/utils"
	"github.com/sirupsen/logrus"
)

type FileReplaceTask struct {
	TypeName string
	Path     string

	Name              string
	Pattern           string
	Repl              string
	Count             int
	AppendIfNotFound  bool
	PrependIfNotFound bool
	NotFoundContent   string
	BackupExtension   string
	MaxFileSize       string

	Require []string

	Creates []string
	OnlyIf  []string
	Unless  []string

	Shell string

	// values created during task build
	MaxFileSizeCalculated uint64
	PatternCompiled       *regexp.Regexp

	// was replace file updated?
	Updated bool
}

type FileReplaceTaskBuilder struct {
}

var fileReplaceTaskParamsFnMap = taskParamsFnMap{
	NameField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.Name = fmt.Sprint(val)
		return nil
	},

	PatternField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		valStr := fmt.Sprint(val)
		t.Pattern = valStr
		return nil
	},
	ReplField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.Repl = fmt.Sprint(val)
		return nil
	},
	CountField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		valStr := fmt.Sprint(val)
		count, err := strconv.Atoi(valStr)
		if err != nil {
			return nil
		}
		t.Count = count
		return nil
	},
	AppendIfNotFoundField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.AppendIfNotFound = conv.ConvertToBool(val)
		return nil
	},
	PrependIfNotFoundField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.PrependIfNotFound = conv.ConvertToBool(val)
		return nil
	},
	NotFoundContentField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.NotFoundContent = fmt.Sprint(val)
		return nil
	},
	BackupExtensionField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		t.BackupExtension = fmt.Sprint(val)
		return nil
	},
	MaxFileSizeField: func(task Task, path string, val interface{}) error {
		t := task.(*FileReplaceTask)
		valStr := fmt.Sprint(val)
		t.MaxFileSize = valStr
		return nil
	},

	CreatesField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*FileReplaceTask)
		t.Creates, err = parseCreatesField(val, path)
		return err
	},
	OnlyIfField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*FileReplaceTask)
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	},
	UnlessField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*FileReplaceTask)
		t.Unless, err = parseUnlessField(val, path)
		return err
	},

	RequireField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*FileReplaceTask)
		t.Require, err = parseRequireField(val, path)
		return err
	},

	ShellField: func(task Task, path string, val interface{}) error {
		valStr := fmt.Sprint(val)
		t := task.(*FileReplaceTask)
		t.Shell = valStr
		return nil
	},
}

func (frtb FileReplaceTaskBuilder) Build(typeName, path string, params interface{}) (t Task, err error) {
	task := &FileReplaceTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := Build(typeName, path, params, task, fileReplaceTaskParamsFnMap)

	return task, errs.ToError()
}

func (t *FileReplaceTask) GetTypeName() string {
	return t.TypeName
}

func (t *FileReplaceTask) GetRequirements() []string {
	return t.Require
}

func (t *FileReplaceTask) GetPath() string {
	return t.Path
}

func (t *FileReplaceTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.TypeName, t.GetPath())
}

func (t *FileReplaceTask) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *FileReplaceTask) GetUnlessCmds() []string {
	return t.Unless
}

func (t *FileReplaceTask) GetCreatesFilesList() []string {
	return t.Creates
}

var (
	ErrAppendAndPrependSetAtTheSameTime = errors.New("append_if_not_found and prepend_if_not_found cannot be set at the same time." +
		"please set one or the other")
)

const (
	defaultMaxFileSize = "512k"
)

func (t *FileReplaceTask) Validate() error {
	errs := &utils.Errors{}

	err := ValidateRequired(t.Name, t.Path+"."+NameField)
	errs.Add(err)

	err = ValidateRequired(t.Pattern, t.Path+"."+PatternField)
	errs.Add(err)

	if len(errs.Errs) > 0 {
		return errs.ToError()
	}

	if t.Pattern != "" {
		compiledRegExp, err := regexp.Compile(t.Pattern)
		if err != nil {
			errs.Add(err)
		}
		t.PatternCompiled = compiledRegExp
	}

	if t.MaxFileSize == "" {
		t.MaxFileSize = defaultMaxFileSize
	}

	MaxFileSizeCalculated, err := conv.ConvertToFileSize(t.MaxFileSize)
	if err != nil {
		errs.Add(err)
	}
	t.MaxFileSizeCalculated = MaxFileSizeCalculated

	if t.AppendIfNotFound && t.PrependIfNotFound {
		errs.Add(ErrAppendAndPrependSetAtTheSameTime)
	}

	return errs.ToError()
}

type FileReplaceTaskExecutor struct {
	FsManager FsManager
	Runner    tacoexec.Runner
}

func (frte *FileReplaceTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := ExecutionResult{
		Changes: make(map[string]string),
	}

	frt, ok := task.(*FileReplaceTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to FileReplaceTask", task)
		return execRes
	}

	execRes.Name = frt.Name

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		Path:         frt.Path,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		Shell:        frt.Shell,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := shouldCheckConditionals(execCtx, frte.FsManager, frte.Runner, frt)
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

	origFilename := frt.Name

	origfileInfo, err := os.Stat(origFilename)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !origfileInfo.Mode().IsRegular() {
		execRes.Err = fmt.Errorf("%s is not a regular file", origFilename)
		return execRes
	}

	if uint64(origfileInfo.Size()) > frt.MaxFileSizeCalculated {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		execRes.SkipReason = "file size is greater than max_file_size"
		return execRes
	}

	start := time.Now()

	rollback := false
	backupFilename := ""
	makeBackup := frt.BackupExtension != ""

	if makeBackup {
		backupFilename = makeBackupFilename(origFilename, frt.BackupExtension)
	}

	origFileContents, err := os.ReadFile(origFilename)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	origfileInfo, err = os.Stat(origFilename)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	var updatedFileContents []byte
	var replacementCount int
	var additionsCount int

	match := frt.PatternCompiled.Match(origFileContents)
	if match {
		updatedFileContents, replacementCount = ReplaceUsingRegexpWithCount(
			origFileContents,
			frt.PatternCompiled,
			frt.Repl,
			frt.Count)
	} else {
		newContent := frt.Repl
		if frt.NotFoundContent != "" {
			newContent = frt.NotFoundContent
		}
		if frt.AppendIfNotFound {
			updatedFileContents = append(origFileContents, []byte(newContent)...) //nolint:gocritic // appendAssign
			additionsCount = 1
		} else if frt.PrependIfNotFound {
			updatedFileContents = append([]byte(newContent), origFileContents...)
			additionsCount = 1
		}
	}

	// will only be non-nil if the original contents have been updated
	if updatedFileContents != nil {
		if makeBackup {
			err := os.WriteFile(backupFilename, origFileContents, origfileInfo.Mode())
			if err != nil {
				execRes.Err = err
				return execRes
			}
			logrus.Debugf("created backup file %s for original file %s", backupFilename, origFilename)
		}

		err = os.WriteFile(frt.Name, updatedFileContents, origfileInfo.Mode())
		if err != nil {
			if !rollback {
				execRes.Err = err
				return execRes
			}
			removeBackupErr := os.Remove(backupFilename)
			if removeBackupErr != nil {
				logrus.Infof("failed to remove backup file: %v", removeBackupErr)
			}
			execRes.Err = err
			return execRes
		}

		frt.Updated = true

		logrus.Debugf("updated file contents for %s", origFilename)
	}

	if frt.Updated && execRes.Err == nil {
		execRes.Comment = "File updated"
		if replacementCount > 0 {
			execRes.Changes["count"] = fmt.Sprintf("%d replacement(s) made", replacementCount)
		} else if additionsCount > 0 {
			execRes.Changes["count"] = fmt.Sprintf("%d addition(s) made", additionsCount)
		}
	}

	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", task.GetPath(), execRes.Duration)
	return execRes
}

func makeBackupFilename(origFilename string, ext string) (backupFilename string) {
	return origFilename + "." + ext
}

func ReplaceUsingRegexpWithCount(contents []byte, re *regexp.Regexp, repl string, maxRepl int) (newContents []byte, replacementCount int) {
	count := 0

	replContents := re.ReplaceAllFunc(contents, func(matchStr []byte) []byte {
		if maxRepl == 0 || count < maxRepl {
			// replace again, this time using matched fragment with replacement using closure captured 're'
			count++
			return re.ReplaceAll(matchStr, []byte(repl))
		}

		// if the max replacements has been reached then just use the original match string without replacement
		return matchStr
	})

	return replContents, count
}
