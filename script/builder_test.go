package script

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"

	"github.com/stretchr/testify/assert"
)

type RawDataProviderMock struct {
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
	return []byte(rdpm.DataToReturn), rdpm.ErrToReturn
}

func (tm *TaskBuilderTaskMock) GetRequirements() []string {
	return tm.Requirements
}

func TestBuilder(t *testing.T) {
	testCases := []struct {
		YamlInput           string
		DataProviderErr     error
		TaskValidationError error
		BuilderError        error
		ExpectedErrMsg      string
		ExpectedScripts     tasks.Scripts
		TaskRequirements    []string
	}{
		{
			YamlInput: `
cwd:
  # Name of the class and the module
  cmd.run:
    - name: echo ${PASSWORD}
    - cwd: /usr/tmp
    - shell: zsh
    - env:
        - PASSWORD: bunny
    - creates: /tmp/my-date.txt
    #- comment: out
    - user: root
    - names:
        - name one
        - name two
        - name three
    - onlyif: echo 1
`,
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
			YamlInput: `
cwd:
  cmd.run:
    - name: echo 1
`,
			BuilderError:    errors.New("failed to build task"),
			ExpectedErrMsg:  "failed to build task",
			ExpectedScripts: tasks.Scripts{},
		},
		{
			YamlInput: `
cwd:
  cmd.run:
    - name: echo 1
    - cwd: /usr/tmp
  somerun:
    - name: echo 2
`,
			TaskValidationError: errors.New("task is invalid"),
			ExpectedErrMsg:      "task is invalid, task is invalid",
			ExpectedScripts:     tasks.Scripts{},
		},
		{
			YamlInput: `
cwd:
  # Name of the class and the module
  cmd.run:
    - names:
						name one
`,
			ExpectedErrMsg:  "yaml: line 6: found character that cannot start any token",
			ExpectedScripts: tasks.Scripts{},
		},
		{
			YamlInput: `
cwd:
  # Name of the class and the module
  cmd.run:
    - names:
        - run one
        - run two
        - run three
    - require:
        - req one
        - req two
        - req three
    - onlyif:
        - onlyif one
        - onlyif two
        - onlyif three
`,
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
			YamlInput: `
manyCreates:
  # Name of the class and the module
  cmd.run:
    - name: many creates cmd
    - require: require one
    - creates:
        - create one
        - create two
        - create three
    - unless: some expected false condition
`,
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
			YamlInput: `
scriptValidation:
  cmd.run:
    - name: task one
    - require: scriptValidation
`,
			ExpectedErrMsg: "task at path 'scriptValidation.cmd.run[1]' cannot require own script 'scriptValidation', " +
				"cyclic requirements are detected: '[scriptValidation]'",
			TaskRequirements: []string{"scriptValidation"},
		},
		{
			YamlInput: `
manyUnless:
  cmd.run:
    - name: expecting for one unless to be false
    - unless:
        - unless one
        - unless two
        - unless three
`,
			ExpectedScripts: tasks.Scripts{
				{
					ID: "manyUnless",
					Tasks: []tasks.Task{
						&TaskBuilderTaskMock{
							TypeName: "cmd.run",
							Path:     "manyUnless.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "expecting for one unless to be false"},
								{tasks.Unless: []interface{}{
									"unless one",
									"unless two",
									"unless three",
								}},
							},
							ValidationError: nil,
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		dataProviderMock := RawDataProviderMock{
			DataToReturn: testCase.YamlInput,
			ErrToReturn:  testCase.DataProviderErr,
		}

		taskBuilderMock := &TaskBuilderMock{
			BuildError:          testCase.BuilderError,
			TaskValidationError: testCase.TaskValidationError,
			TaskRequirements:    testCase.TaskRequirements,
		}

		parser := Builder{
			DataProvider: dataProviderMock,
			TaskBuilder:  taskBuilderMock,
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
