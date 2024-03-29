package filereplace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/realvnc-labs/tacoscript/conv"
	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/shared/conditionals"
	"github.com/realvnc-labs/tacoscript/tasks/shared/executionresult"
	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/sirupsen/logrus"
)

var (
	ErrAppendAndPrependSetAtTheSameTime = errors.New("append_if_not_found and prepend_if_not_found cannot be set at the same time." +
		"please set one or the other")
)

const (
	TaskType = "file.replace"

	defaultMaxFileSize = "512k"
)

type Task struct {
	TypeName string // TaskType
	Path     string // TaskName

	Name              string   `taco:"name"` // Target
	Pattern           string   `taco:"pattern"`
	Repl              string   `taco:"repl"`
	Count             int      `taco:"count"`
	AppendIfNotFound  bool     `taco:"append_if_not_found"`
	PrependIfNotFound bool     `taco:"prepend_if_not_found"`
	NotFoundContent   string   `taco:"not_found_content"`
	BackupExtension   string   `taco:"backup"`
	MaxFileSize       string   `taco:"max_file_size"`
	Require           []string `taco:"require"`
	Creates           []string `taco:"creates"`
	OnlyIf            []string `taco:"onlyif"`
	Unless            []string `taco:"unless"`
	Shell             string   `taco:"shell"`

	// values created during task build
	maxFileSizeCalculated uint64
	patternCompiled       *regexp.Regexp

	// was replace file updated?
	Updated bool
}

func (t *Task) GetTypeName() string {
	return t.TypeName
}

func (t *Task) GetRequirements() []string {
	return t.Require
}

func (t *Task) GetPath() string {
	return t.Path
}

func (t *Task) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.TypeName, t.GetPath())
}

func (t *Task) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *Task) GetUnlessCmds() []string {
	return t.Unless
}

func (t *Task) GetCreatesFilesList() []string {
	return t.Creates
}

func (t *Task) Validate(goos string) error {
	errs := &utils.Errors{}

	err := tasks.ValidateRequired(t.Name, t.Path+"."+tasks.NameField)
	errs.Add(err)

	err = tasks.ValidateRequired(t.Pattern, t.Path+"."+tasks.PatternField)
	errs.Add(err)

	if len(errs.Errs) > 0 {
		return errs.ToError()
	}

	if t.Pattern != "" {
		compiledRegExp, err := regexp.Compile(t.Pattern)
		if err != nil {
			errs.Add(err)
		}
		t.patternCompiled = compiledRegExp
	}

	if t.MaxFileSize == "" {
		t.MaxFileSize = defaultMaxFileSize
	}

	MaxFileSizeCalculated, err := conv.ConvertToFileSize(t.MaxFileSize)
	if err != nil {
		errs.Add(err)
	}
	t.maxFileSizeCalculated = MaxFileSizeCalculated

	if t.AppendIfNotFound && t.PrependIfNotFound {
		errs.Add(ErrAppendAndPrependSetAtTheSameTime)
	}

	return errs.ToError()
}

type Executor struct {
	FsManager tasks.FsManager
	Runner    tacoexec.Runner
}

func (frte *Executor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := executionresult.ExecutionResult{
		Changes: make(map[string]string),
	}

	frt, ok := task.(*Task)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to Task", task)
		return execRes
	}

	execRes.Name = frt.Name
	execRes.Comment = "File not changed"

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		Path:         frt.Path,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		Shell:        frt.Shell,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := conditionals.Check(execCtx, frte.FsManager, frte.Runner, frt)
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

	origfileInfo, err := frte.FsManager.Stat(origFilename)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !origfileInfo.Mode().IsRegular() {
		execRes.Err = fmt.Errorf("%s is not a regular file", origFilename)
		return execRes
	}

	if uint64(origfileInfo.Size()) > frt.maxFileSizeCalculated {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		execRes.SkipReason = "file size is greater than max_file_size"
		return execRes
	}

	start := time.Now()

	backupFilename := ""
	makeBackup := frt.BackupExtension != ""

	if makeBackup {
		backupFilename = makeBackupFilename(origFilename, frt.BackupExtension)
	}

	origFileContents, err := frte.FsManager.ReadFile(origFilename)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	var updatedFileContents string
	var replacementCount int
	var additionsCount int

	match := frt.patternCompiled.MatchString(origFileContents)
	if match {
		updatedFileContents, replacementCount = ReplaceUsingRegexpWithCount(
			origFileContents,
			frt.patternCompiled,
			frt.Repl,
			frt.Count)
	} else {
		newContent := frt.Repl
		if frt.NotFoundContent != "" {
			newContent = frt.NotFoundContent
		}
		additionsCount = 1
		if frt.AppendIfNotFound {
			updatedFileContents = origFileContents + newContent
		} else if frt.PrependIfNotFound {
			updatedFileContents = newContent + origFileContents
		}
	}

	// will only be non-nil if the original contents have been updated
	if updatedFileContents != "" {
		if makeBackup {
			err := frte.FsManager.WriteFile(backupFilename, origFileContents, origfileInfo.Mode())
			if err != nil {
				execRes.Err = err
				return execRes
			}
			logrus.Debugf("created backup file %s for original file %s", backupFilename, origFilename)
		}

		err := frte.FsManager.WriteFile(frt.Name, updatedFileContents, origfileInfo.Mode())
		if err != nil {
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

func ReplaceUsingRegexpWithCount(contents string, re *regexp.Regexp, repl string, maxRepl int) (newContents string, replacementCount int) {
	count := 0

	replContents := re.ReplaceAllStringFunc(contents, func(matchStr string) string {
		if maxRepl == 0 || count < maxRepl {
			count++
			// replace again, this time using matched fragment with replacement using closure captured 're'
			// this ensures that any reg exp group replacements or similar are applied
			return re.ReplaceAllString(matchStr, repl)
		}

		// if the max replacements has been reached then just use the original match string without replacement
		return matchStr
	})

	return replContents, count
}
