package script

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
	"github.com/realvnc-labs/tacoscript/utils"
	"gopkg.in/yaml.v2"

	"github.com/realvnc-labs/tacoscript/tasks"

	"github.com/stretchr/testify/assert"
)

type RawDataProviderMock struct {
	FileName     string
	DataToReturn string
	ErrToReturn  error
}

type TaskBuilderMock struct {
	BuildError          error
	TaskValidationError error
	TaskRequirements    []string
}

func (bm *TaskBuilderMock) Build(typeName, path string, ctx interface{}) (tasks.CoreTask, error) {
	t := &TaskBuilderTaskMock{
		TypeName:        typeName,
		Path:            path,
		Context:         ctx,
		ValidationError: bm.TaskValidationError,
		Requirements:    bm.TaskRequirements,
	}

	return t, bm.BuildError
}

type TaskBuilderTaskMock struct {
	TypeName        string
	Path            string
	Context         interface{}
	ValidationError error
	Requirements    []string
	OnlyIf          []string
	Unless          []string
	Creates         []string
}

func (tm *TaskBuilderTaskMock) GetTypeName() string {
	return tm.TypeName
}

func (tm *TaskBuilderTaskMock) Execute(ctx context.Context) executionresult.ExecutionResult {
	return executionresult.ExecutionResult{}
}

func (tm *TaskBuilderTaskMock) Validate(goos string) error {
	return tm.ValidationError
}

func (tm *TaskBuilderTaskMock) GetPath() string {
	return tm.Path
}

func (rdpm RawDataProviderMock) Read() ([]byte, error) {
	if rdpm.FileName != "" {
		return os.ReadFile(filepath.Join("yaml", rdpm.FileName))
	}

	return []byte(rdpm.DataToReturn), rdpm.ErrToReturn
}

func (tm *TaskBuilderTaskMock) GetRequirements() []string {
	return tm.Requirements
}

func (tm *TaskBuilderTaskMock) GetOnlyIfCmds() []string {
	return tm.OnlyIf
}

func (tm *TaskBuilderTaskMock) GetUnlessCmds() []string {
	return tm.Unless
}

func (tm *TaskBuilderTaskMock) GetCreatesFilesList() []string {
	return tm.Creates
}

type TemplateVariablesProviderMock struct {
	Variables              utils.TemplateVarsMap
	TemplateVariablesError error
}

func (tvpm TemplateVariablesProviderMock) GetTemplateVariables() (utils.TemplateVarsMap, error) {
	return tvpm.Variables, tvpm.TemplateVariablesError
}

func TestBuilder(t *testing.T) {
	testCases := []struct {
		YamlFileName           string
		YamlInput              string
		DataProviderErr        error
		TaskValidationError    error
		BuilderError           error
		ExpectedErrMsg         string
		ExpectedScripts        tasks.Scripts
		TaskRequirements       []string
		TemplateVariables      utils.TemplateVarsMap
		TemplateVariablesError error
	}{
		{
			YamlFileName: "test1.yaml",
			ExpectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "cwd",
					Tasks: []tasks.CoreTask{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "cwd.cmd.run[1]",
							Context: []interface{}{
								yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "echo ${PASSWORD}"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.CwdField, Value: "/usr/tmp"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "zsh"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.EnvField, Value: []interface{}{
									yaml.MapSlice{yaml.MapItem{Key: "PASSWORD", Value: "bunny"}},
								}}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "/tmp/my-date.txt"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.UserField, Value: "root"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.NamesField, Value: []interface{}{
									"name one",
									"name two",
									"name three",
								}}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: "echo 1"}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			DataProviderErr: errors.New("data not available"),
			ExpectedErrMsg:  "data not available",
			ExpectedScripts: tasks.Scripts{},
		},
		{
			YamlFileName:    "test2.yaml",
			BuilderError:    errors.New("failed to build task"),
			ExpectedErrMsg:  "failed to build task",
			ExpectedScripts: tasks.Scripts{},
		},
		{
			YamlFileName:        "test3.yaml",
			TaskValidationError: errors.New("task is invalid"),
			ExpectedErrMsg:      "task is invalid, task is invalid",
			ExpectedScripts:     tasks.Scripts{},
		},
		{
			YamlFileName:    "test4.yaml",
			ExpectedErrMsg:  "invalid script provided: yaml: line 5: found character that cannot start any token",
			ExpectedScripts: tasks.Scripts{},
		},
		{
			YamlFileName: "test5.yaml",
			ExpectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "cwd",
					Tasks: []tasks.CoreTask{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "cwd.cmd.run[1]",
							Context: []interface{}{
								yaml.MapSlice{yaml.MapItem{Key: tasks.NamesField, Value: []interface{}{
									"run one",
									"run two",
									"run three",
								}}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
									"req one",
									"req two",
									"req three",
								}}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: []interface{}{
									"onlyif one",
									"onlyif two",
									"onlyif three",
								}}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			YamlFileName: "test6.yaml",
			ExpectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "manyCreates",
					Tasks: []tasks.CoreTask{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "manyCreates.cmd.run[1]",
							Context: []interface{}{
								yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "many creates cmd"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: "require one"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
									"create one",
									"create two",
									"create three",
								}}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: "some expected false condition"}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			YamlFileName: "test7.yaml",
			ExpectedErrMsg: "task at path 'scriptValidation.cmd.run[1]' cannot require own script 'scriptValidation', " +
				"cyclic requirements are detected: '[scriptValidation]'",
			TaskRequirements: []string{"scriptValidation"},
		},
		{
			YamlFileName: "test9.go.yaml",
			TemplateVariables: utils.TemplateVarsMap{
				utils.OSFamily: "RedHat",
			},
			ExpectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "template",
					Tasks: []tasks.CoreTask{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "template.cmd.run[1]",
							Context: []interface{}{
								yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "yum --version"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
									"test.txt",
								}}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			YamlFileName: "test9.go.yaml",
			TemplateVariables: utils.TemplateVarsMap{
				utils.OSFamily: "Ubuntu",
			},
			ExpectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "template",
					Tasks: []tasks.CoreTask{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "template.cmd.run[1]",
							Context: []interface{}{
								yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "apt --version"}},
								yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
									"test.txt",
								}}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			TemplateVariables: utils.TemplateVarsMap{
				utils.OSFamily: "",
			},
			YamlFileName:           "test9.go.yaml",
			ExpectedErrMsg:         "cannot provide template variables",
			ExpectedScripts:        tasks.Scripts{},
			TemplateVariablesError: errors.New("cannot provide template variables"),
		},
		{
			YamlFileName:    "test10.go.yaml",
			ExpectedErrMsg:  `template: goyaml:3:6: executing "goyaml" at <eq "RedHat">: error calling eq: missing argument for comparison`,
			ExpectedScripts: tasks.Scripts{},
		},
	}

	for _, testCase := range testCases {
		dataProviderMock := RawDataProviderMock{
			DataToReturn: testCase.YamlInput,
			ErrToReturn:  testCase.DataProviderErr,
			FileName:     testCase.YamlFileName,
		}

		taskBuilderMock := &TaskBuilderMock{
			BuildError:          testCase.BuilderError,
			TaskValidationError: testCase.TaskValidationError,
			TaskRequirements:    testCase.TaskRequirements,
		}

		templateVariablesProviderMock := TemplateVariablesProviderMock{
			Variables:              testCase.TemplateVariables,
			TemplateVariablesError: testCase.TemplateVariablesError,
		}

		parser := Builder{
			DataProvider:              dataProviderMock,
			TaskBuilder:               taskBuilderMock,
			TemplateVariablesProvider: templateVariablesProviderMock,
		}

		scripts, err := parser.BuildScripts()
		if testCase.ExpectedErrMsg != "" {
			assert.EqualError(t, err, testCase.ExpectedErrMsg, testCase.ExpectedErrMsg)
			continue
		}

		assert.NoError(t, err)
		if err != nil {
			continue
		}

		assert.EqualValues(t, testCase.ExpectedScripts, scripts)
	}
}
