package tasks

import (
	"context"
	"errors"
	exec2 "github.com/cloudradar-monitoring/tacoscript/exec"
	"os/exec"
	"strings"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/stretchr/testify/assert"
)

func TestTaskExecution(t *testing.T) {
	systemAPIMock := &exec2.SystemAPIMock{
		Cmds: []*exec.Cmd{},
	}

	runnerMock := &exec2.SystemRunner{SystemAPI: systemAPIMock}

	testCases := []struct {
		Task            *CmdRunTask
		ExpectedResult  ExecutionResult
		RunnerMock      *exec2.SystemRunner
		ExpectedCmdStrs []string
		Name            string
	}{
		{
			Name: "test one name command with 2 envs",
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
				FsManager:             &utils.FsManagerMock{},
				Runner:                runnerMock,
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "some std out",
				StdErr:    "some std err",
			},
			ExpectedCmdStrs: []string{"zsh -c some test command"},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				ErrToGive:          nil,
				StdErrText:         "some std err",
				StdOutText:         "some std out",
				UserSetErrToReturn: nil,
			}},
		},
		{
			Name: "test skip command if file exists",
			Task: &CmdRunTask{
				User:                  "some user",
				Name:                  "some parser command",
				MissingFilesCondition: []string{"somefile.txt"},
				FsManager: &utils.FsManagerMock{
					ExistsToReturn: true,
				},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				ErrToGive:          nil,
				UserSetErrToReturn: errors.New("some error"),
			}},
		},
		{
			Name: "test setting user failure",
			Task: &CmdRunTask{
				Name:      "echo 12345",
				User:      "some user",
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some error"),
			},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				ErrToGive:          nil,
				UserSetErrToReturn: errors.New("some error"),
			}},
		},
		{
			Name: "same cmd execution failure",
			Task: &CmdRunTask{
				Name:      "lpwd",
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some runner error"),
			},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: errors.New("some runner error"),
			}},
		},
		{
			Name: "execute multiple names",
			Task: &CmdRunTask{
				Path: "many names path",
				Names: []string{
					"many names cmd 1",
					"many names cmd 2",
					"many names cmd 3",
				},
				WorkingDir: "/many/dev",
				User:       "usermany",
				FsManager:  &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"many names cmd 1", "many names cmd 2", "many names cmd 3"},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "test multiple create file conditions",
			Task: &CmdRunTask{
				Name: "cmd with many MissingFilesConditions",
				MissingFilesCondition: []string{
					"file.one",
					"file.two",
				},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"cmd with many MissingFilesConditions"},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "executing one onlyif condition with success",
			Task: &CmdRunTask{
				Name:      "cmd lala",
				OnlyIf:    []string{"check before lala"},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"check before lala", "cmd lala"},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "executing one onlyif condition with failure",
			Task: &CmdRunTask{
				Name:      "cmd with OnlyIf failure",
				OnlyIf:    []string{"check OnlyIf error"},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "cmd with OnlyIf failure") {
						return nil
					}

					return errors.New("some OnlyIfFailure")
				},
			},
			}},
		{
			Name: "executing multiple onlyif conditions with failure",
			Task: &CmdRunTask{
				Name:      "cmd with multiple OnlyIf failure",
				OnlyIf:    []string{"check OnlyIf success", "check OnlyIf failure"},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "check OnlyIf failure") {
						return errors.New("check OnlyIf failure")
					}

					return nil
				},
			}},
		},
		{
			Name: "executing multiple onlyif conditions with success",
			Task: &CmdRunTask{
				Name:      "cmd with multiple OnlyIf success",
				OnlyIf:    []string{"check OnlyIf success 1", "check OnlyIf success 2"},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"check OnlyIf success 1", "check OnlyIf success 2", "cmd with multiple OnlyIf success"},
			RunnerMock: &exec2.SystemRunner{SystemAPI: &exec2.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			testCase.Task.Runner = testCase.RunnerMock

			res := testCase.Task.Execute(context.Background())
			assert.EqualValues(t, testCase.ExpectedResult.Err, res.Err)
			assert.EqualValues(t, testCase.ExpectedResult.IsSkipped, res.IsSkipped)
			assert.EqualValues(t, testCase.ExpectedResult.StdOut, res.StdOut)
			assert.EqualValues(t, testCase.ExpectedResult.StdErr, res.StdErr)


			if testCase.ExpectedResult.Err != nil {
				return
			}

			systemAPIMock := testCase.RunnerMock.SystemAPI.(*exec2.SystemAPIMock)
			cmds := systemAPIMock.Cmds

			AssertCmdsPartiallyMatch(t, testCase.ExpectedCmdStrs, cmds)
			if testCase.ExpectedResult.IsSkipped {
				return
			}

			assert.Equal(t, len(testCase.ExpectedCmdStrs), len(cmds))
			for _, cmd := range cmds {
				assert.Equal(t, testCase.Task.WorkingDir, cmd.Dir)
				AssertEnvValuesMatch(t, testCase.Task.Envs, cmd.Env)
			}

			assert.Equal(t, testCase.Task.User, systemAPIMock.UserNameInput)
			assert.Equal(t, testCase.Task.Path, systemAPIMock.UserNamePathInput)
		})
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

func TestOSCmdRunnerValidation(t *testing.T) {
	testCases := []struct {
		Task          CmdRunTask
		ExpectedError string
	}{
		{
			Task: CmdRunTask{
				Names:     []string{"one", "two"},
				Errors:    &utils.Errors{},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Name:      "three",
				Errors:    &utils.Errors{},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Name:      "four",
				Names:     []string{"five", "six"},
				Errors:    &utils.Errors{},
				FsManager: &utils.FsManagerMock{},
			},
			ExpectedError: "",
		},
		{
			Task:          CmdRunTask{Errors: &utils.Errors{}, FsManager: &utils.FsManagerMock{}},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
		{
			Task: CmdRunTask{
				Names:     []string{"", ""},
				Errors:    &utils.Errors{},
				FsManager: &utils.FsManagerMock{},
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
