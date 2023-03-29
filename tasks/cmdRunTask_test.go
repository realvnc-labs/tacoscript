package tasks

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/realvnc-labs/tacoscript/apptest"
	"github.com/realvnc-labs/tacoscript/conv"

	appExec "github.com/realvnc-labs/tacoscript/exec"
)

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		Task            *CmdRunTask
		ExpectedResult  ExecutionResult
		RunnerMock      *appExec.SystemRunner
		Name            string
		FileShouldExist bool
	}{
		{
			Name: "test one name command with 2 envs",
			Task: &CmdRunTask{
				Path:       "somepath",
				Named:      NamedTask{Name: "some test command"},
				WorkingDir: "/tmp/dev",
				User:       "user",
				Shell:      "zsh",
				Envs: conv.KeyValues{
					{
						Key:   "someenvkey1",
						Value: "someenvval2",
					},
				},
				Creates: []string{""},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "some std out",
				StdErr:    "some std err",
			},
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
				User:    "some user",
				Named:   NamedTask{Name: "some parser command"},
				Creates: []string{"somefile.txt"},
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
			FileShouldExist: true,
		},
		{
			Name: "test setting user failure",
			Task: &CmdRunTask{
				Named: NamedTask{Name: "echo 12345"},
				User:  "some user",
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
				Named: NamedTask{Name: "lpwd"},
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
				Named: NamedTask{Names: []string{
					"many names cmd 1",
					"many names cmd 2",
					"many names cmd 3",
				}},
				WorkingDir: "/many/dev",
				User:       "usermany",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "test multiple create file conditions",
			Task: &CmdRunTask{
				Named: NamedTask{Name: "cmd with many MissingFilesConditions"},
				Creates: []string{
					"file.one",
					"file.two",
				},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "executing one onlyif condition with success",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd lala"},
				OnlyIf: []string{"check before lala"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:      []*exec.Cmd{},
				ErrToGive: nil,
			}},
		},
		{
			Name: "executing one onlyif condition with failure",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd with OnlyIf failure"},
				OnlyIf: []string{"check OnlyIf error"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
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
			},
		},
		{
			Name: "executing multiple onlyif conditions with failure",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd with multiple OnlyIf failure"},
				OnlyIf: []string{"check OnlyIf success", "check OnlyIf failure"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
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
				Named:  NamedTask{Name: "cmd with multiple OnlyIf success"},
				OnlyIf: []string{"check OnlyIf success 1", "check OnlyIf success 2"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing onlyif validation error",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd"},
				OnlyIf: []string{"checking onlyif validation error"},
				User:   "some user 123",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("cannot set user"),
			},
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
		{
			Name: "executing one unless condition with success",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd masa"},
				Unless: []string{"run unless masa"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "run unless masa") {
						return appExec.RunError{Err: errors.New("run unless masa failed")}
					}

					return nil
				},
			}},
		},
		{
			Name: "executing one unless condition with failure",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd with unless failure"},
				Unless: []string{"check unless failure"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing multiple unless conditions with all success",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd with multiple unless success"},
				Unless: []string{"check unless one", "check unless two"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing multiple unless conditions with at least one failure",
			Task: &CmdRunTask{
				Named:  NamedTask{Name: "cmd with multiple unless with at least one failure"},
				Unless: []string{"check unless 1", "check unless 2"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					if strings.Contains(cmd.String(), "check unless 2") {
						return appExec.RunError{Err: errors.New("check unless 2 failed")}
					}
					return nil
				},
			}},
		},
		{
			Name: "executing unless validation error",
			Task: &CmdRunTask{
				Unless: []string{"checking unless validation error"},
				Named:  NamedTask{Name: "executing unless validation error"},
				User:   "some user 345",
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				UserSetErrToReturn: errors.New("cannot set user 345"),
			}},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("cannot set user 345"),
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			cmdRunExecutor := &CmdRunTaskExecutor{
				Runner: tc.RunnerMock,
				FsManager: &apptest.FsManagerMock{
					FileExistsExistsToReturn: tc.FileShouldExist,
				},
			}

			res := cmdRunExecutor.Execute(context.Background(), tc.Task)
			assert.EqualValues(tt, tc.ExpectedResult.Err, res.Err)
			assert.EqualValues(tt, tc.ExpectedResult.IsSkipped, res.IsSkipped)
			assert.EqualValues(tt, tc.ExpectedResult.StdOut, res.StdOut)
			assert.EqualValues(tt, tc.ExpectedResult.StdErr, res.StdErr)

			if tc.ExpectedResult.Err != nil {
				return
			}

			systemAPIMock := tc.RunnerMock.SystemAPI.(*appExec.SystemAPIMock)
			cmds := systemAPIMock.Cmds

			if tc.ExpectedResult.IsSkipped {
				return
			}

			for _, cmd := range cmds {
				assert.Equal(tt, tc.Task.WorkingDir, cmd.Dir)
				AssertEnvValuesMatch(tt, tc.Task.Envs, cmd.Env)
			}

			assert.Equal(tt, tc.Task.User, systemAPIMock.UserNameInput)
			assert.Equal(tt, tc.Task.Path, systemAPIMock.UserNamePathInput)
		})
	}
}

func TestCmdRunTaskValidation(t *testing.T) {
	testCases := []struct {
		Task          CmdRunTask
		ExpectedError string
	}{
		{
			Task: CmdRunTask{
				Named: NamedTask{Names: []string{"one", "two"}},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Named: NamedTask{Name: "three"},
			},
			ExpectedError: "",
		},
		{
			Task: CmdRunTask{
				Named: NamedTask{Names: []string{"five", "six"}, Name: "four"},
			},
			ExpectedError: "",
		},
		{
			Task:          CmdRunTask{},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
		{
			Task: CmdRunTask{
				Named: NamedTask{Names: []string{"", ""}},
			},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
	}

	for _, testCase := range testCases {
		err := testCase.Task.Validate(runtime.GOOS)
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
	}
}
