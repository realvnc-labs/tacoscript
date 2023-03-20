package tasks

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/sirupsen/logrus"
)

type CmdRunTask struct {
	TypeName string
	Path     string
	Named    NamedTask
	Envs     conv.KeyValues

	WorkingDir string   `taco:"cwd"`
	User       string   `taco:"user"`
	Shell      string   `taco:"shell"`
	Creates    []string `taco:"creates"`
	Require    []string `taco:"require"`
	OnlyIf     []string `taco:"onlyif"`
	Unless     []string `taco:"unless"`

	mapper FieldNameMapper

	// aborts task execution if one task fails
	AbortOnError bool
}

type CmdRunTaskBuilder struct {
}

var cmdRunTaskParamsFnMap = taskFieldsParserConfig{
	NameField: {
		parseFn: func(task Task, path string, val interface{}) error {
			t := task.(*CmdRunTask)
			t.Named.Name = fmt.Sprint(val)
			return nil
		},
		fieldName: "Name",
	},
	NamesField: {
		parseFn: func(task Task, path string, val interface{}) error {
			var err error
			t := task.(*CmdRunTask)
			t.Named.Names, err = conv.ConvertToValues(val)
			return err
		},
		fieldName: "Names",
	},
	EnvField: {
		parseFn: func(task Task, path string, val interface{}) error {
			var err error
			t := task.(*CmdRunTask)
			t.Envs, err = conv.ConvertToKeyValues(val, path)
			return err
		},
		fieldName: "Env",
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

func (crt *CmdRunTask) Validate(goos string) error {
	errs := &utils.Errors{}
	err1 := ValidateRequired(crt.Named.Name, crt.Path+"."+NameField)
	err2 := ValidateRequiredMany(crt.Named.Names, crt.Path+"."+NamesField)

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

func (crt *CmdRunTask) GetMapper() (mapper FieldNameMapper) {
	if crt.mapper == nil {
		crt.mapper = newFieldNameMapper()
	}
	return crt.mapper
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
	execRes.Name = strings.Join(cmdRunTask.Named.GetNames(), "; ")

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		WorkingDir:   cmdRunTask.WorkingDir,
		User:         cmdRunTask.User,
		Path:         cmdRunTask.Path,
		Envs:         cmdRunTask.Envs,
		Cmds:         cmdRunTask.Named.GetNames(),
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
	logrus.Debugf("execution of %s has finished, took: %v", cmdRunTask.Named.Name, execRes.Duration)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()
	execRes.Pid = execCtx.Pid

	return execRes
}
