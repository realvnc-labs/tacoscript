package tasks

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"

	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/realvnc-labs/tacoscript/conv"

	"github.com/sirupsen/logrus"
)

type CmdRunTask struct {
	TypeName string
	Path     string
	NamedTask
	WorkingDir string
	User       string
	Shell      string
	Envs       conv.KeyValues
	Creates    []string
	Require    []string
	OnlyIf     []string
	Unless     []string

	// aborts task execution if one task fails
	AbortOnError bool
}

type CmdRunTaskBuilder struct {
}

var cmdRunTaskParamsFnMap = taskParamsFnMap{
	NameField: func(task Task, path string, val interface{}) error {
		t := task.(*CmdRunTask)
		t.Name = fmt.Sprint(val)
		return nil
	},
	CwdField: func(task Task, path string, val interface{}) error {
		t := task.(*CmdRunTask)
		t.WorkingDir = fmt.Sprint(val)
		return nil
	},
	UserField: func(task Task, path string, val interface{}) error {
		t := task.(*CmdRunTask)
		t.User = fmt.Sprint(val)
		return nil
	},
	EnvField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.Envs, err = conv.ConvertToKeyValues(val, path)
		return err
	},
	NamesField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.Names, err = conv.ConvertToValues(val, path)
		return err
	},
	AbortOnErrorField: func(task Task, path string, val interface{}) error {
		t := task.(*CmdRunTask)
		t.AbortOnError = conv.ConvertToBool(val)
		return nil
	},

	CreatesField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.Creates, err = parseCreatesField(val, path)
		return err
	},
	OnlyIfField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.OnlyIf, err = parseOnlyIfField(val, path)
		return err
	},
	UnlessField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.Unless, err = parseUnlessField(val, path)
		return err
	},

	RequireField: func(task Task, path string, val interface{}) error {
		var err error
		t := task.(*CmdRunTask)
		t.Require, err = parseRequireField(val, path)
		return err
	},

	ShellField: func(task Task, path string, val interface{}) error {
		valStr := fmt.Sprint(val)
		t := task.(*CmdRunTask)
		t.Shell = valStr
		return nil
	},
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, params interface{}) (t Task, err error) {
	task := &CmdRunTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := Build(typeName, path, params, task, cmdRunTaskParamsFnMap)

	return task, errs.ToError()
}

func (crt *CmdRunTask) GetTypeName() string {
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

func (crt *CmdRunTask) GetOnlyIfCmds() []string {
	return crt.OnlyIf
}

func (crt *CmdRunTask) GetUnlessCmds() []string {
	return crt.Unless
}

func (crt *CmdRunTask) GetCreatesFilesList() []string {
	return crt.Creates
}

type CmdRunTaskExecutor struct {
	Runner    tacoexec.Runner
	FsManager FsManager
}

func (crte *CmdRunTaskExecutor) Execute(ctx context.Context, task Task) ExecutionResult {
	execRes := ExecutionResult{}
	cmdRunTask, ok := task.(*CmdRunTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to CmdRunTask", task)
		return execRes
	}
	execRes.Name = strings.Join(cmdRunTask.GetNames(), "; ")

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   cmdRunTask.WorkingDir,
		User:         cmdRunTask.User,
		Path:         cmdRunTask.Path,
		Envs:         cmdRunTask.Envs,
		Cmds:         cmdRunTask.GetNames(),
		Shell:        cmdRunTask.Shell,
	}

	shouldNotBeExecutedReason, err := checkConditionals(execCtx, crte.FsManager, crte.Runner, cmdRunTask)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if shouldNotBeExecutedReason != "" {
		execRes.IsSkipped = true
		execRes.SkipReason = shouldNotBeExecutedReason
		execRes.Comment = `Command "` + execRes.Name + `" did not run: ` + shouldNotBeExecutedReason
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
	execRes.Pid = execCtx.Pid

	return execRes
}
