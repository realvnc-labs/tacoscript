package tasks

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/stretchr/testify/assert"
)

func TestTaskExecution(t *testing.T) {
	systemAPIMock := &appExec.SystemAPIMock{
		Cmds: []*exec.Cmd{},
	}

	runnerMock := &appExec.SystemRunner{SystemAPI: systemAPIMock}

	testCases := []struct {
		Task            *CmdRunTask
		ExpectedResult  ExecutionResult
		RunnerMock      *appExec.SystemRunner
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
				Err:       appExec.RunError{Err: errors.New("some runner error")},
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "cmd with OnlyIf failure") {
						return nil
					}

					return appExec.RunError{Err: errors.New("some OnlyIfFailure")}
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "check OnlyIf failure") {
						return appExec.RunError{Err: errors.New("check OnlyIf failure")}
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
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing onlyif validation error",
			Task: &CmdRunTask{
				Name:      "executing onlyif validation error",
				OnlyIf:    []string{"checking onlyif validation error"},
				FsManager: &utils.FsManagerMock{},
				User:      "some user 123",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("cannot set user"),
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				UserSetErrToReturn: errors.New("cannot set user"),
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "checking onlyif validation error") {
						return errors.New("onlyIf validation failure")
					}
					return nil
				},
			}},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			tc.Task.Runner = tc.RunnerMock

			res := tc.Task.Execute(context.Background())
			assert.EqualValues(tt, tc.ExpectedResult.Err, res.Err)
			assert.EqualValues(tt, tc.ExpectedResult.IsSkipped, res.IsSkipped)
			assert.EqualValues(tt, tc.ExpectedResult.StdOut, res.StdOut)
			assert.EqualValues(tt, tc.ExpectedResult.StdErr, res.StdErr)

			if tc.ExpectedResult.Err != nil {
				return
			}

			systemAPIMock := tc.RunnerMock.SystemAPI.(*appExec.SystemAPIMock)
			cmds := systemAPIMock.Cmds

			AssertCmdsPartiallyMatch(tt, tc.ExpectedCmdStrs, cmds)
			if tc.ExpectedResult.IsSkipped {
				return
			}

			assert.Equal(tt, len(tc.ExpectedCmdStrs), len(cmds))
			for _, cmd := range cmds {
				assert.Equal(tt, tc.Task.WorkingDir, cmd.Dir)
				AssertEnvValuesMatch(tt, tc.Task.Envs, cmd.Env)
			}

			assert.Equal(tt, tc.Task.User, systemAPIMock.UserNameInput)
			assert.Equal(tt, tc.Task.Path, systemAPIMock.UserNamePathInput)
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
