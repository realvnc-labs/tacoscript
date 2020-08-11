package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/stretchr/testify/assert"
)

type OsExecutorMock struct {
	stdOutText string
	stdErrText string
	cmds       []*exec.Cmd
	errToGive  error
	id         string

	userNameInput      string
	userNamePathInput  string
	userSetErrToReturn error
}

func (oem *OsExecutorMock) Run(cmd *exec.Cmd) error {
	_, err := cmd.Stdout.Write([]byte(oem.stdOutText))
	if err != nil {
		return err
	}

	_, err = cmd.Stderr.Write([]byte(oem.stdErrText))
	if err != nil {
		return err
	}

	oem.cmds = append(oem.cmds, cmd)

	return oem.errToGive
}

func (oem *OsExecutorMock) SetUser(userName, path string, cmd *exec.Cmd) error {
	oem.userNameInput = userName
	oem.userNamePathInput = path

	return oem.userSetErrToReturn
}

func TestCmdRunTaskBuilder(t *testing.T) {
	runnerMock := &OsExecutorMock{
		cmds: []*exec.Cmd{},
		id:   "some id",
	}

	testCases := []struct {
		typeName     string
		path         string
		ctx          []map[string]interface{}
		expectedTask *CmdRunTask
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
				},
			},
			expectedTask: &CmdRunTask{
				TypeName:   "someType",
				Path:       "somePath",
				Name:       "1",
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
				OsExecutor:            runnerMock,
				Errors:                &ValidationErrors{},
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
				TypeName:   "someTypeWithErrors",
				Path:       "somePathWithErrors",
				Envs:       conv.KeyValues{},
				OsExecutor: runnerMock,
				Errors: &ValidationErrors{
					Errs: []error{
						fmt.Errorf("key value array expected at 'somePathWithErrors' but got '123'"),
					},
				},
			},
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
				TypeName:   "someTypeWithErrors2",
				Path:       "somePathWithErrors2",
				Envs:       conv.KeyValues{},
				OsExecutor: runnerMock,
				Errors: &ValidationErrors{
					Errs: []error{
						errors.New(`wrong key value element at 'somePathWithErrors2': '"one"'`),
					},
				},
			},
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
				Names: []string{
					"name one",
					"name two",
				},
				OsExecutor: runnerMock,
				Errors:     &ValidationErrors{},
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
				},
			},
			expectedTask: &CmdRunTask{
				TypeName:   "manyCreatesType",
				Path:       "manyCreatesPath",
				Name:       "many creates command",
				OsExecutor: runnerMock,
				Errors:     &ValidationErrors{},
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
			},
		},
	}

	for _, testCase := range testCases {
		cmdBuilder := CmdRunTaskBuilder{
			OsExecutor: runnerMock,
		}
		actualTask, err := cmdBuilder.Build(
			testCase.typeName,
			testCase.path,
			testCase.ctx,
		)

		assert.NoError(t, err)
		if err != nil {
			continue
		}

		actualCmdRunTask, ok := actualTask.(*CmdRunTask)
		assert.True(t, ok)
		if !ok {
			continue
		}

		assert.Equal(t, testCase.expectedTask.User, actualCmdRunTask.User)
		AssertEnvValuesMatch(t, testCase.expectedTask.Envs, actualCmdRunTask.Envs.ToEqualSignStrings())
		assert.Equal(t, testCase.expectedTask.Path, actualCmdRunTask.Path)
		assert.Equal(t, testCase.expectedTask.WorkingDir, actualCmdRunTask.WorkingDir)
		assert.Equal(t, testCase.expectedTask.MissingFilesCondition, actualCmdRunTask.MissingFilesCondition)
		assert.Equal(t, testCase.expectedTask.Name, actualCmdRunTask.Name)
		assert.Equal(t, testCase.expectedTask.TypeName, actualCmdRunTask.TypeName)
		assert.Equal(t, testCase.expectedTask.Shell, actualCmdRunTask.Shell)
		assert.Equal(t, testCase.expectedTask.Names, actualCmdRunTask.Names)
		assert.Equal(t, testCase.expectedTask.Require, actualCmdRunTask.Require)
		assert.EqualValues(t, testCase.expectedTask.Errors, actualCmdRunTask.Errors)
	}
}

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		Task                            *CmdRunTask
		ExpectedResult                  ExecutionResult
		RunnerMock                      *OsExecutorMock
		ExpectedCmdStrs                 []string
		ShouldCreateFileForMissingCheck bool
	}{
		{
			Task: &CmdRunTask{
				Path:       "somepath",
				Name:       "some test command",
				WorkingDir: "/tmp/dev",
				User:       "user",
				Shell:      "zsh",
				Envs: conv.KeyValues{
					{
						Key:   "someenvkey1",
						Value: "someenvval2",
					},
				},
				MissingFilesCondition: []string{""},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "some std out",
				StdErr:    "some std err",
			},
			ExpectedCmdStrs: []string{"zsh -c some test command"},
			RunnerMock: &OsExecutorMock{
				cmds:               []*exec.Cmd{},
				errToGive:          nil,
				id:                 "some id",
				stdErrText:         "some std err",
				stdOutText:         "some std out",
				userSetErrToReturn: nil,
			},
		},
		{
			Task: &CmdRunTask{
				User:                  "some user",
				Name:                  "some parser command",
				MissingFilesCondition: []string{"somefile.txt"},
			},
			ShouldCreateFileForMissingCheck: true,
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &OsExecutorMock{
				cmds:               []*exec.Cmd{},
				errToGive:          nil,
				id:                 "some id",
				userSetErrToReturn: errors.New("some error"),
			},
		},
		{
			Task: &CmdRunTask{
				Name: "echo 12345",
				User: "some user",
			},
			ShouldCreateFileForMissingCheck: false,
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some error"),
			},
			RunnerMock: &OsExecutorMock{
				cmds:               []*exec.Cmd{},
				errToGive:          nil,
				id:                 "some id",
				userSetErrToReturn: errors.New("some error"),
			},
		},
		{
			Task: &CmdRunTask{
				Name: "lpwd",
			},
			ShouldCreateFileForMissingCheck: false,
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some runner error"),
			},
			RunnerMock: &OsExecutorMock{
				cmds:      []*exec.Cmd{},
				errToGive: errors.New("some runner error"),
				id:        "some id",
			},
		},
		{
			Task: &CmdRunTask{
				Path: "many names path",
				Names: []string{
					"many names cmd 1",
					"many names cmd 2",
					"many names cmd 3",
				},
				WorkingDir: "/many/dev",
				User:       "usermany",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"many names cmd 1", "many names cmd 2", "many names cmd 3"},
			RunnerMock: &OsExecutorMock{
				cmds:      []*exec.Cmd{},
				errToGive: nil,
				id:        "some id",
			},
		},
		{
			Task: &CmdRunTask{
				Name: "cmd with many MissingFilesConditions",
				MissingFilesCondition: []string{
					"file.one",
					"file.two",
				},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"cmd with many MissingFilesConditions"},
			RunnerMock: &OsExecutorMock{
				cmds:      []*exec.Cmd{},
				errToGive: nil,
				id:        "some id",
			},
			ShouldCreateFileForMissingCheck: false,
		},
	}

	filesToDelete := make([]string, 0)
	defer func() {
		if len(filesToDelete) == 0 {
			return
		}
		for _, file := range filesToDelete {
			err := os.Remove(file)
			if err != nil {
				log.Warn(err)
			}
		}
	}()

	for _, testCase := range testCases {
		testCase.Task.OsExecutor = testCase.RunnerMock

		if testCase.ShouldCreateFileForMissingCheck && len(testCase.Task.MissingFilesCondition) > 0 {
			for _, fileToCreate := range testCase.Task.MissingFilesCondition {
				emptyFile, err := os.Create(fileToCreate)
				assert.NoError(t, err)
				if err != nil {
					continue
				}

				err = emptyFile.Close()
				assert.NoError(t, err)
				filesToDelete = append(filesToDelete, fileToCreate)
			}
		}

		res := testCase.Task.Execute(context.Background())
		assert.EqualValues(t, testCase.ExpectedResult.Err, res.Err)
		assert.EqualValues(t, testCase.ExpectedResult.IsSkipped, res.IsSkipped)
		assert.EqualValues(t, testCase.ExpectedResult.StdOut, res.StdOut)
		assert.EqualValues(t, testCase.ExpectedResult.StdErr, res.StdErr)

		cmds := testCase.RunnerMock.cmds

		if testCase.ExpectedResult.IsSkipped {
			assert.Len(t, cmds, 0)
			continue
		}

		if testCase.ExpectedResult.Err != nil {
			continue
		}

		AssertCmdsPartiallyMatch(t, testCase.ExpectedCmdStrs, cmds)

		assert.Equal(t, len(testCase.ExpectedCmdStrs), len(cmds))
		for _, cmd := range cmds {
			assert.Equal(t, testCase.Task.WorkingDir, cmd.Dir)
			AssertEnvValuesMatch(t, testCase.Task.Envs, cmd.Env)
		}

		assert.Equal(t, testCase.Task.User, testCase.RunnerMock.userNameInput)
		assert.Equal(t, testCase.Task.Path, testCase.RunnerMock.userNamePathInput)
	}
}

func AssertCmdsPartiallyMatch(t *testing.T, expectedCmds []string, actualExecutedCmds []*exec.Cmd) {
	notFoundCmds := make([]string, 0, len(expectedCmds))

	executedCmdStrs := make([]string, 0, len(actualExecutedCmds))
	for _, actualCmd := range actualExecutedCmds {
		executedCmdStrs = append(executedCmdStrs, actualCmd.String())
	}

	for _, expectedCmdStr := range expectedCmds {
		cmdFound := false
		for _, actualCmdStr := range executedCmdStrs {
			if strings.HasSuffix(actualCmdStr, expectedCmdStr) {
				cmdFound = true
				break
			}
		}
		if !cmdFound {
			notFoundCmds = append(notFoundCmds, expectedCmdStr)
		}
	}

	assert.Empty(
		t,
		notFoundCmds,
		"was not able to find following expected commands '%s' in the list of executed commands '%s'",
		strings.Join(notFoundCmds, ", "),
		strings.Join(executedCmdStrs, ", "),
	)
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

func TestOSCmdRunnerValidation(t *testing.T) {
	testCases := []struct {
		Task          CmdRunTask
		ExpectedError string
	}{
		{
			Task: CmdRunTask{
				Names:  []string{"one", "two"},
				Errors: &ValidationErrors{},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Name:   "three",
				Errors: &ValidationErrors{},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Name:   "four",
				Names:  []string{"five", "six"},
				Errors: &ValidationErrors{},
			},
			ExpectedError: "",
		},
		{
			Task:          CmdRunTask{Errors: &ValidationErrors{}},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
		{
			Task: CmdRunTask{
				Names:  []string{"", ""},
				Errors: &ValidationErrors{},
			},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
	}

	for _, testCase := range testCases {
		err := testCase.Task.Validate()
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
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
