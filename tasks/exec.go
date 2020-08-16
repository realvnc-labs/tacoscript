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
	Errors                *utils.Errors
	Runner                exec2.Runner
	OnlyIf                []string
	Unless                []string
	FsManager             utils.FsManager
}

type CmdRunTaskBuilder struct {
	Runner    exec2.Runner
	FsManager utils.FsManager
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, ctx []map[string]interface{}) (Task, error) {
	t := &CmdRunTask{
		TypeName:  typeName,
		Path:      path,
		Errors:    &utils.Errors{},
		Runner:    crtb.Runner,
		FsManager: crtb.FsManager,
	}

	for _, contextItem := range ctx {
		for key, val := range contextItem {
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
				envs, err := conv.ConvertToKeyValues(val, path)
				t.Errors.Add(err)
				t.Envs = envs
			case CreatesField:
				crtb.parseCreatesField(t, val, path)
			case NamesField:
				names, err := conv.ConvertToValues(val, path)
				t.Errors.Add(err)
				t.Names = names
			case RequireField:
				crtb.parseRequireField(t, val, path)
			case OnlyIf:
				crtb.parseOnlyIfField(t, val, path)
			case Unless:
				crtb.parseUnlessField(t, val, path)
			}
		}
	}

	return t, nil
}

func (crtb CmdRunTaskBuilder) parseRequireField(t *CmdRunTask, val interface{}, path string) {
	requireItems := make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		requireItems = append(requireItems, typedVal)
	default:
		requireItems, err = conv.ConvertToValues(val, path)
		t.Errors.Add(err)
	}
	t.Require = requireItems
}

func (crtb CmdRunTaskBuilder) parseCreatesField(t *CmdRunTask, val interface{}, path string) {
	createsItems := make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		createsItems = append(createsItems, typedVal)
	default:
		createsItems, err = conv.ConvertToValues(val, path)
		t.Errors.Add(err)
	}
	t.MissingFilesCondition = createsItems
}

func (crtb CmdRunTaskBuilder) parseOnlyIfField(t *CmdRunTask, val interface{}, path string) {
	onlyIf := make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		onlyIf = append(onlyIf, typedVal)
	default:
		onlyIf, err = conv.ConvertToValues(val, path)
		t.Errors.Add(err)
	}
	t.OnlyIf = onlyIf
}

func (crtb CmdRunTaskBuilder) parseUnlessField(t *CmdRunTask, val interface{}, path string) {
	unless := make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		unless = append(unless, typedVal)
	default:
		unless, err = conv.ConvertToValues(val, path)
		t.Errors.Add(err)
	}
	t.Unless = unless
}

func (crt *CmdRunTask) GetName() string {
	return crt.TypeName
}

func (crt *CmdRunTask) GetRequirements() []string {
	return crt.Require
}

func (crt *CmdRunTask) Validate() error {
	err1 := ValidateRequired(crt.Name, crt.Path+"."+NameField)
	err2 := ValidateRequiredMany(crt.Names, crt.Path+"."+NamesField)

	if err1 != nil && err2 != nil {
		crt.Errors.Add(err1)
		crt.Errors.Add(err2)
		return crt.Errors.ToError()
	}

	return nil
}

func (crt *CmdRunTask) GetPath() string {
	return crt.Path
}

func (crt *CmdRunTask) Execute(ctx context.Context) ExecutionResult {
	execRes := ExecutionResult{}

	rawCmds := []string{crt.Name}
	rawCmds = append(rawCmds, crt.Names...)

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &exec2.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   crt.WorkingDir,
		User:         crt.User,
		Path:         crt.Path,
		Envs:         crt.Envs,
		Cmds:         rawCmds,
		Shell:        crt.Shell,
	}

	shouldBeExecuted, err := crt.shouldBeExecuted(execCtx)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if !shouldBeExecuted {
		logrus.Info("command will be skipped")
		execRes.IsSkipped = true
		return execRes
	}

	start := time.Now()

	err = crt.Runner.Run(execCtx)

	execRes.Duration = time.Since(start)
	if err != nil {
		execRes.Err = err
	}
	logrus.Debugf("execution of %s has finished, took: %v", crt.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()

	return execRes
}

func (crt *CmdRunTask) checkOnlyIfs(ctx *exec2.Context) (isSuccess bool, err error) {
	if len(crt.OnlyIf) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = crt.OnlyIf
	err = crt.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Infof("will skip cmd since onlyif condition has failed: %v", runErr)
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (crt *CmdRunTask) checkUnless(ctx *exec2.Context) (isExpectationSuccess bool, err error) {
	if len(crt.Unless) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = crt.Unless

	err = crt.Runner.Run(&newCtx)

	if err != nil {
		runErr, isRunErr := err.(exec2.RunError)
		if isRunErr {
			logrus.Infof("will continue cmd since at least one unless condition has failed: %v", runErr)
			return true, nil
		}

		return false, err
	}

	logrus.Infof("will stop cmd since all unless conditions din't fail")
	return false, nil
}

func (crt *CmdRunTask) shouldBeExecuted(ctx *exec2.Context) (shouldBeExecuted bool, err error) {
	isExists, err := crt.checkMissingFileCondition()
	if err != nil {
		return false, err
	}

	if isExists {
		return false, nil
	}

	isSuccess, err := crt.checkOnlyIfs(ctx)
	if err != nil {
		return false, err
	}

	if !isSuccess {
		return false, nil
	}

	isExpectationSuccess, err := crt.checkUnless(ctx)
	if err != nil {
		return false, err
	}

	if !isExpectationSuccess {
		return false, nil
	}

	return true, nil
}

func (crt *CmdRunTask) checkMissingFileCondition() (isExists bool, err error) {
	if len(crt.MissingFilesCondition) == 0 {
		return
	}

	for _, missingFileCondition := range crt.MissingFilesCondition {
		if missingFileCondition == "" {
			continue
		}
		isExists, err = crt.FsManager.FileExists(missingFileCondition)
		if err != nil {
			err = fmt.Errorf("failed to check if file '%s' exists: %w", missingFileCondition, err)
			return
		}

		if isExists {
			logrus.Infof("file '%s' exists, will skip command '%s'", missingFileCondition, crt.Name)
			return
		}
	}

	return
}
