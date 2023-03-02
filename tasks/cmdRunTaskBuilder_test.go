package tasks

import (
	"strings"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/conv"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestCmdRunTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []interface{}
		expectedTask  *CmdRunTask
		expectedError string
	}{
		{
			typeName: "someType",
			path:     "somePath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "1"}},
				yaml.MapSlice{yaml.MapItem{Key: CwdField, Value: "somedir"}},
				yaml.MapSlice{yaml.MapItem{Key: UserField, Value: "someuser"}},
				yaml.MapSlice{yaml.MapItem{Key: ShellField, Value: "someshell"}},
				yaml.MapSlice{yaml.MapItem{Key: EnvField, Value: []interface{}{
					yaml.MapSlice{yaml.MapItem{Key: "one", Value: "1"}},
					yaml.MapSlice{yaml.MapItem{Key: "two", Value: "2"}},
				}}},
				yaml.MapSlice{yaml.MapItem{Key: CreatesField, Value: "somefile.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: "one condition"}},
			},
			expectedTask: &CmdRunTask{
				TypeName:   "someType",
				Path:       "somePath",
				NamedTask:  NamedTask{Name: "1"},
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
				yaml.MapSlice{yaml.MapItem{Key: EnvField, Value: 123}},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors",
				Path:     "somePathWithErrors",
				Envs:     conv.KeyValues{},
			},
			expectedError: "key value array expected at 'somePathWithErrors' but got '123'",
		},
		{
			typeName: "someTypeWithErrors2",
			path:     "somePathWithErrors2",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: EnvField, Value: []interface{}{
					"one",
				}}},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors2",
				Path:     "somePathWithErrors2",
				Envs:     conv.KeyValues{},
			},
			expectedError: `wrong key value element at 'somePathWithErrors2': '"one"'`,
		},
		{
			typeName: "manyNamesType",
			path:     "manyNamesPath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: "one require field"}},
				yaml.MapSlice{yaml.MapItem{Key: NamesField, Value: []interface{}{
					"name one",
					"name two",
				}}},
			},
			expectedTask: &CmdRunTask{
				TypeName: "manyNamesType",
				Path:     "manyNamesPath",
				Require: []string{
					"one require field",
				},
				NamedTask: NamedTask{Names: []string{
					"name one",
					"name two",
				}},
			},
		},
		{
			typeName: "manyCreatesType",
			path:     "manyCreatesPath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "many creates command"}},
				yaml.MapSlice{yaml.MapItem{Key: CreatesField, Value: []interface{}{
					"create one",
					"create two",
					"create three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
			},
			expectedTask: &CmdRunTask{
				TypeName:  "manyCreatesType",
				Path:      "manyCreatesPath",
				NamedTask: NamedTask{Name: "many creates command"},
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
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "one unless value"}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: "unless one"}},
			},
			expectedTask: &CmdRunTask{
				TypeName:  "oneUnlessValue",
				Path:      "oneUnlessValuePath",
				NamedTask: NamedTask{Name: "one unless value"},
				Unless: []string{
					"unless one",
				},
			},
		},
		{
			typeName: "manyUnlessValue",
			path:     "manyUnlessValuePath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "many unless value"}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: []interface{}{
					"Unless one",
					"Unless two",
					"Unless three",
				}}},
			},
			expectedTask: &CmdRunTask{
				TypeName:  "manyUnlessValue",
				Path:      "manyUnlessValuePath",
				NamedTask: NamedTask{Name: "many unless value"},
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
			cmdBuilder := CmdRunTaskBuilder{}
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

			actualCmdRunTask, ok := actualTask.(*CmdRunTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assert.Equal(t, tc.expectedTask.User, actualCmdRunTask.User)
			AssertEnvValuesMatch(t, tc.expectedTask.Envs, actualCmdRunTask.Envs.ToEqualSignStrings())
			assert.Equal(t, tc.expectedTask.Path, actualCmdRunTask.Path)
			assert.Equal(t, tc.expectedTask.WorkingDir, actualCmdRunTask.WorkingDir)
			assert.Equal(t, tc.expectedTask.Creates, actualCmdRunTask.Creates)
			assert.Equal(t, tc.expectedTask.Name, actualCmdRunTask.Name)
			assert.Equal(t, tc.expectedTask.TypeName, actualCmdRunTask.TypeName)
			assert.Equal(t, tc.expectedTask.Shell, actualCmdRunTask.Shell)
			assert.Equal(t, tc.expectedTask.Names, actualCmdRunTask.Names)
			assert.Equal(t, tc.expectedTask.Require, actualCmdRunTask.Require)
			assert.Equal(t, tc.expectedTask.OnlyIf, actualCmdRunTask.OnlyIf)
			assert.Equal(t, tc.expectedTask.Unless, actualCmdRunTask.Unless)
		})
	}
}

func AssertEnvValuesMatch(t *testing.T, expectedEnvs conv.KeyValues, actualCmdEnvs []string) {
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
