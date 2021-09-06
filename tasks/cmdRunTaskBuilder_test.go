package tasks

import (
	"strings"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/conv"

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
			ctx: []map[string]interface{}{
				{
					NameField:  1,
					CwdField:   "somedir",
					UserField:  "someuser",
					ShellField: "someshell",
					EnvField: BuildExpectedEnvs(map[interface{}]interface{}{
						"one": "1",
						"two": "2",
					}),
					CreatesField: "somefile.txt",
					OnlyIf:       "one condition",
				},
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
				MissingFilesCondition: []string{"somefile.txt"},
				OnlyIf:                []string{"one condition"},
			},
		},
		{
			typeName: "someTypeWithErrors",
			path:     "somePathWithErrors",
			ctx: []map[string]interface{}{
				{
					EnvField: 123,
				},
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
			ctx: []map[string]interface{}{
				{
					EnvField: []interface{}{
						"one",
					},
				},
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
			ctx: []map[string]interface{}{
				{
					RequireField: "one require field",
					NamesField: []interface{}{
						"name one",
						"name two",
					},
				},
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
			ctx: []map[string]interface{}{
				{
					NameField: "many creates command",
					CreatesField: []interface{}{
						"create one",
						"create two",
						"create three",
					},
					RequireField: []interface{}{
						"req one",
						"req two",
						"req three",
					},
					OnlyIf: []interface{}{
						"OnlyIf one",
						"OnlyIf two",
						"OnlyIf three",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName:  "manyCreatesType",
				Path:      "manyCreatesPath",
				NamedTask: NamedTask{Name: "many creates command"},
				MissingFilesCondition: []string{
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
			ctx: []map[string]interface{}{
				{
					NameField: "one unless value",
					Unless:    "unless one",
				},
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
			ctx: []map[string]interface{}{
				{
					NameField: "many unless value",
					Unless: []interface{}{
						"Unless one",
						"Unless two",
						"Unless three",
					},
				},
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
			assert.Equal(t, tc.expectedTask.MissingFilesCondition, actualCmdRunTask.MissingFilesCondition)
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

func BuildExpectedEnvs(expectedEnvs map[interface{}]interface{}) []interface{} {
	envs := make([]interface{}, 0, len(expectedEnvs))
	for envKey, envValue := range expectedEnvs {
		envs = append(envs, map[interface{}]interface{}{
			envKey: envValue,
		})
	}

	return envs
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
