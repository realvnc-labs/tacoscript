package parse

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

type ParserBuilderMock struct {
	BuildError          error
	TaskValidationError error
}

func (bm *ParserBuilderMock) Build(typeName, path string, ctx []map[string]interface{}) (tasks.Task, error) {
	t := TaskMock{
		TypeName:        typeName,
		Path:            path,
		Context:         ctx,
		ValidationError: bm.TaskValidationError,
	}

	return t, bm.BuildError
}

type TaskMock struct {
	TypeName        string
	Path            string
	Context         []map[string]interface{}
	ValidationError error
}

func (tm TaskMock) GetName() string {
	return tm.TypeName
}

func (tm TaskMock) Execute(ctx context.Context) tasks.ExecutionResult {
	return tasks.ExecutionResult{}
}

func (tm TaskMock) Validate() error {
	return tm.ValidationError
}

func (tm TaskMock) GetPath() string {
	return tm.Path
}

func (rdpm RawDataProviderMock) Read() ([]byte, error) {
	return []byte(rdpm.DataToReturn), rdpm.ErrToReturn
}

func TestYamlParser(t *testing.T) {
	testCases := []struct {
		YamlInput           string
		DataProviderErr     error
		TaskValidationError error
		BuilderError        error
		ExpectedErrMsg      string
		ExpectedScripts     tasks.Scripts
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
`,
			ExpectedScripts: tasks.Scripts{
				{
					ID: "cwd",
					Tasks: []tasks.Task{
						TaskMock{
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
`,
			ExpectedScripts: tasks.Scripts{
				{
					ID: "cwd",
					Tasks: []tasks.Task{
						TaskMock{
							TypeName: "cmd.run",
							Path:     "cwd.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NamesField: []interface{}{
									"run one",
									"run two",
									"run three",
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
    - creates:
        - create one
        - create two
        - create three
`,
			ExpectedScripts: tasks.Scripts{
				{
					ID: "manyCreates",
					Tasks: []tasks.Task{
						TaskMock{
							TypeName: "cmd.run",
							Path:     "manyCreates.cmd.run[1]",
							Context: []map[string]interface{}{
								{tasks.NameField: "many creates cmd"},
								{tasks.CreatesField: []interface{}{
									"create one",
									"create two",
									"create three",
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

		taskBuilderMock := &ParserBuilderMock{
			BuildError:          testCase.BuilderError,
			TaskValidationError: testCase.TaskValidationError,
		}

		parser := Parser{
			DataProvider: dataProviderMock,
			TaskBuilder:  taskBuilderMock,
		}

		scripts, err := parser.ParseScripts()
		if testCase.ExpectedErrMsg != "" {
			assert.EqualError(t, err, testCase.ExpectedErrMsg)
			continue
		}

		assert.NoError(t, err)
		if err != nil {
			continue
		}

		assert.Equal(t, testCase.ExpectedScripts, scripts)
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
