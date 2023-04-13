package cmdrun

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
	"github.com/realvnc-labs/tacoscript/tasks/shared/executionresult"
	"github.com/realvnc-labs/tacoscript/tasks/shared/names"

	appExec "github.com/realvnc-labs/tacoscript/exec"
)

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		InputTask       *Task
		ExpectedResult  executionresult.ExecutionResult
		RunnerMock      *appExec.SystemRunner
		Name            string
		FileShouldExist bool
	}{
		{
			Name: "test one name command with 2 envs",
			InputTask: &Task{
				Path:       "somepath",
				Named:      names.TaskNames{Name: "some test command"},
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
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				User:    "some user",
				Named:   names.TaskNames{Name: "some parser command"},
				Creates: []string{"somefile.txt"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named: names.TaskNames{Name: "echo 12345"},
				User:  "some user",
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named: names.TaskNames{Name: "lpwd"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Path: "many names path",
				Named: names.TaskNames{Names: []string{
					"many names cmd 1",
					"many names cmd 2",
					"many names cmd 3",
				}},
				WorkingDir: "/many/dev",
				User:       "usermany",
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named: names.TaskNames{Name: "cmd with many MissingFilesConditions"},
				Creates: []string{
					"file.one",
					"file.two",
				},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd lala"},
				OnlyIf: []string{"check before lala"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with OnlyIf failure"},
				OnlyIf: []string{"check OnlyIf error"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with multiple OnlyIf failure"},
				OnlyIf: []string{"check OnlyIf success", "check OnlyIf failure"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with multiple OnlyIf success"},
				OnlyIf: []string{"check OnlyIf success 1", "check OnlyIf success 2"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing onlyif validation error",
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd"},
				OnlyIf: []string{"checking onlyif validation error"},
				User:   "some user 123",
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd masa"},
				Unless: []string{"run unless masa"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with unless failure"},
				Unless: []string{"check unless failure"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing multiple unless conditions with all success",
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with multiple unless success"},
				Unless: []string{"check unless one", "check unless two"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
		},
		{
			Name: "executing multiple unless conditions with at least one failure",
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with multiple unless with at least one failure"},
				Unless: []string{"check unless 1", "check unless 2"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			InputTask: &Task{
				Unless: []string{"checking unless validation error"},
				Named:  names.TaskNames{Name: "executing unless validation error"},
				User:   "some user 345",
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds:               []*exec.Cmd{},
				UserSetErrToReturn: errors.New("cannot set user 345"),
			}},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("cannot set user 345"),
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			cmdRunExecutor := &CrtExecutor{
				Runner: tc.RunnerMock,
				FsManager: &apptest.FsManagerMock{
					FileExistsExistsToReturn: tc.FileShouldExist,
				},
			}

			res := cmdRunExecutor.Execute(context.Background(), tc.InputTask)
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
				assert.Equal(tt, tc.InputTask.WorkingDir, cmd.Dir)
				assertEnvValuesMatch(tt, tc.InputTask.Envs, cmd.Env)
			}

			assert.Equal(tt, tc.InputTask.User, systemAPIMock.UserNameInput)
			assert.Equal(tt, tc.InputTask.Path, systemAPIMock.UserNamePathInput)
		})
	}
}

func TestCmdRunTaskValidation(t *testing.T) {
	testCases := []struct {
		InputTask     Task
		ExpectedError string
	}{
		{
			InputTask: Task{
				Named: names.TaskNames{Names: []string{"one", "two"}},
			},
			ExpectedError: "",
		},
		{
			InputTask: Task{
				Named: names.TaskNames{Name: "three"},
			},
			ExpectedError: "",
		},
		{
			InputTask: Task{
				Named: names.TaskNames{Names: []string{"five", "six"}, Name: "four"},
			},
			ExpectedError: "",
		},
		{
			InputTask:     Task{},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
		{
			InputTask: Task{
				Named: names.TaskNames{Names: []string{"", ""}},
			},
			ExpectedError: "empty required value at path '.name', empty required values at path '.names'",
		},
	}

	for _, testCase := range testCases {
		err := testCase.InputTask.Validate(runtime.GOOS)
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
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
