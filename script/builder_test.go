package script

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/tasks"

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

func (bm *TaskBuilderMock) Build(typeName, path string, ctx []map[string]interface{}) (tasks.Task, error) {
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
	Context         []map[string]interface{}
	ValidationError error
	Requirements    []string
}

func (tm *TaskBuilderTaskMock) GetName() string {
	return tm.TypeName
}

func (tm *TaskBuilderTaskMock) Execute(ctx context.Context) tasks.ExecutionResult {
	return tasks.ExecutionResult{}
}

func (tm *TaskBuilderTaskMock) Validate() error {
	return tm.ValidationError
}

func (tm *TaskBuilderTaskMock) GetPath() string {
	return tm.Path
}

func (rdpm RawDataProviderMock) Read() ([]byte, error) {
	if rdpm.FileName != "" {
		return ioutil.ReadFile("yaml" + string(os.PathSeparator) + rdpm.FileName)
	}

	return []byte(rdpm.DataToReturn), rdpm.ErrToReturn
}

func (tm *TaskBuilderTaskMock) GetRequirements() []string {
	return tm.Requirements
}

type TemplateVariablesProviderMock struct {
	Variables              map[string]interface{}
	TemplateVariablesError error
}

func (tvpm TemplateVariablesProviderMock) GetTemplateVariables() (map[string]interface{}, error) {
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
		TemplateVariables      map[string]interface{}
		TemplateVariablesError error
	}{
		{
			YamlFileName: "test1.yaml",
			ExpectedScripts: tasks.Scripts{
				{
					ID: "cwd",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "cwd.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "echo ${PASSWORD}"},
								{tasks.CwdField: "/usr/tmp"},
								{tasks.ShellField: "zsh"},
								{
									tasks.EnvField: BuildExpectedEnvs(map[interface{}]interface{}{
										"PASSWORD": "bunny",
									}),
								},
								{tasks.CreatesField: "/tmp/my-date.txt"},
								{tasks.UserField: "root"},
								{tasks.NamesField: []interface{}{
									"name one",
									"name two",
									"name three",
								}},
								{tasks.OnlyIf: "echo 1"},
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
				{
					ID: "cwd",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "cwd.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NamesField: []interface{}{
									"run one",
									"run two",
									"run three",
								}},
								{tasks.RequireField: []interface{}{
									"req one",
									"req two",
									"req three",
								}},
								{tasks.OnlyIf: []interface{}{
									"onlyif one",
									"onlyif two",
									"onlyif three",
								}},
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
				{
					ID: "manyCreates",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "manyCreates.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "many creates cmd"},
								{tasks.RequireField: "require one"},
								{tasks.CreatesField: []interface{}{
									"create one",
									"create two",
									"create three",
								}},
								{tasks.Unless: "some expected false condition"},
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
			TemplateVariables: map[string]interface{}{
				utils.OSFamily: "RedHat",
			},
			ExpectedScripts: tasks.Scripts{
				{
					ID: "template",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "template.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "yum --version"},
								{tasks.CreatesField: []interface{}{
									"test.txt",
								}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			YamlFileName: "test9.go.yaml",
			TemplateVariables: map[string]interface{}{
				utils.OSFamily: "Ubuntu",
			},
			ExpectedScripts: tasks.Scripts{
				{
					ID: "template",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "template.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "apt --version"},
								{tasks.CreatesField: []interface{}{
									"test.txt",
								}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
		{
			TemplateVariables: map[string]interface{}{
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

func BuildExpectedEnvs(expectedEnvs map[interface{}]interface{}) []interface{} {
	envs := make([]interface{}, 0, len(expectedEnvs))
	for envKey, envValue := range expectedEnvs {
		envs = append(envs, map[interface{}]interface{}{
			envKey: envValue,
		})
	}

	return envs
}
