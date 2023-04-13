package crtbuilder

import (
	"strings"
	"testing"

	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
	"github.com/realvnc-labs/tacoscript/tasks/shared/names"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []interface{}
		expectedTask  *cmdrun.Task
		expectedError string
	}{
		{
			typeName: "someType",
			path:     "somePath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "1"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CwdField, Value: "somedir"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UserField, Value: "someuser"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.EnvField, Value: []interface{}{
					yaml.MapSlice{yaml.MapItem{Key: "one", Value: "1"}},
					yaml.MapSlice{yaml.MapItem{Key: "two", Value: "2"}},
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "somefile.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: "one condition"}},
			},
			expectedTask: &cmdrun.Task{
				TypeName:   "someType",
				Path:       "somePath",
				Named:      names.TaskNames{Name: "1"},
				WorkingDir: "somedir",
				User:       "someuser",
				Shell:      "someshell",
				Envs: conv.KeyValues{
					{
						Key:   "one",
						Value: "1",
					},
					{
						Key:   "two",
						Value: "2",
					},
				},
				Creates: []string{"somefile.txt"},
				OnlyIf:  []string{"one condition"},
			},
		},

		{
			typeName: "someTypeWithErrors",
			path:     "somePathWithErrors",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.EnvField, Value: 123}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "someTypeWithErrors",
				Path:     "somePathWithErrors",
				Envs:     conv.KeyValues{},
			},
			expectedError: "key value array expected at 'somePathWithErrors' but got '123': env",
		},
		{
			typeName: "someTypeWithErrors2",
			path:     "somePathWithErrors2",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.EnvField, Value: []interface{}{
					"one",
				}}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "someTypeWithErrors2",
				Path:     "somePathWithErrors2",
				Envs:     conv.KeyValues{},
			},
			expectedError: `wrong key value element at 'somePathWithErrors2': '"one"': env`,
		},
		{
			typeName: "manyNamesType",
			path:     "manyNamesPath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: "one require field"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.NamesField, Value: []interface{}{
					"name one",
					"name two",
				}}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "manyNamesType",
				Path:     "manyNamesPath",
				Require: []string{
					"one require field",
				},
				Named: names.TaskNames{Names: []string{
					"name one",
					"name two",
				}},
			},
		},
		{
			typeName: "manyCreatesType",
			path:     "manyCreatesPath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "many creates command"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
					"create one",
					"create two",
					"create three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "manyCreatesType",
				Path:     "manyCreatesPath",
				Named:    names.TaskNames{Name: "many creates command"},
				Creates: []string{
					"create one",
					"create two",
					"create three",
				},
				Require: []string{
					"req one",
					"req two",
					"req three",
				},
				OnlyIf: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
			},
		},
		{
			typeName: "oneUnlessValue",
			path:     "oneUnlessValuePath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "one unless value"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: "unless one"}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "oneUnlessValue",
				Path:     "oneUnlessValuePath",
				Named:    names.TaskNames{Name: "one unless value"},
				Unless: []string{
					"unless one",
				},
			},
		},
		{
			typeName: "manyUnlessValue",
			path:     "manyUnlessValuePath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "many unless value"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"Unless one",
					"Unless two",
					"Unless three",
				}}},
			},
			expectedTask: &cmdrun.Task{
				TypeName: "manyUnlessValue",
				Path:     "manyUnlessValuePath",
				Named:    names.TaskNames{Name: "many unless value"},
				Unless: []string{
					"Unless one",
					"Unless two",
					"Unless three",
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			cmdBuilder := TaskBuilder{}
			actualTask, err := cmdBuilder.Build(
				tc.typeName,
				tc.path,
				tc.ctx,
			)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			if err != nil {
				return
			}

			actualCmdRunTask, ok := actualTask.(*cmdrun.Task)
			assert.True(t, ok)
			if !ok {
				return
			}

			assert.Equal(t, tc.expectedTask.User, actualCmdRunTask.User)
			assertEnvValuesMatch(t, tc.expectedTask.Envs, actualCmdRunTask.Envs.ToEqualSignStrings())
			assert.Equal(t, tc.expectedTask.Path, actualCmdRunTask.Path)
			assert.Equal(t, tc.expectedTask.WorkingDir, actualCmdRunTask.WorkingDir)
			assert.Equal(t, tc.expectedTask.Creates, actualCmdRunTask.Creates)
			assert.Equal(t, tc.expectedTask.Named, actualCmdRunTask.Named)
			assert.Equal(t, tc.expectedTask.TypeName, actualCmdRunTask.TypeName)
			assert.Equal(t, tc.expectedTask.Shell, actualCmdRunTask.Shell)
			assert.Equal(t, tc.expectedTask.Require, actualCmdRunTask.Require)
			assert.Equal(t, tc.expectedTask.OnlyIf, actualCmdRunTask.OnlyIf)
			assert.Equal(t, tc.expectedTask.Unless, actualCmdRunTask.Unless)
		})
	}
}

func assertEnvValuesMatch(t *testing.T, expectedEnvs conv.KeyValues, actualCmdEnvs []string) {
	expectedRawEnvs := expectedEnvs.ToEqualSignStrings()
	notFoundEnvs := make([]string, 0, len(expectedEnvs))
	for _, expectedRawEnv := range expectedRawEnvs {
		foundEnv := false
		for _, actualCmdEnv := range actualCmdEnvs {
			if expectedRawEnv == actualCmdEnv {
				foundEnv = true
				break
			}
		}

		if !foundEnv {
			notFoundEnvs = append(notFoundEnvs, expectedRawEnv)
		}
	}

	assert.Empty(
		t,
		notFoundEnvs,
		"was not able to find expected environment variables %s in cmd envs %s",
		strings.Join(notFoundEnvs, ", "),
		strings.Join(actualCmdEnvs, ", "),
	)
}
