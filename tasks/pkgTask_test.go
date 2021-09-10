package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/stretchr/testify/assert"
)

type PackageManagerMock struct {
	givenCtx     context.Context
	givenTask    *PkgTask
	outputToGive *PackageManagerExecutionResult
	errToGive    error
}

func (pmm *PackageManagerMock) ExecuteTask(ctx context.Context, t *PkgTask) (res *PackageManagerExecutionResult, err error) {
	pmm.givenCtx = ctx
	pmm.givenTask = t

	return pmm.outputToGive, pmm.errToGive
}

func TestPkgTaskValidation(t *testing.T) {
	testCases := []struct {
		Name          string
		ExpectedError string
		Task          PkgTask
	}{
		{
			Name: "missing_name_and_names",
			Task: PkgTask{
				Path:       "somepath",
				ActionType: ActionUpdate,
			},
			ExpectedError: fmt.Sprintf(
				"empty required value at path 'somepath.%s', empty required values at path 'somepath.%s'",
				NameField,
				NamesField,
			),
		},
		{
			Name: "valid_task_name",
			Task: PkgTask{
				NamedTask:  NamedTask{Name: "some name"},
				ActionType: ActionUninstall,
			},
			ExpectedError: "",
		},
		{
			Name: "valid_task_names",
			Task: PkgTask{
				NamedTask:  NamedTask{Names: []string{"some name1", "some name 2"}},
				ActionType: ActionInstall,
			},
			ExpectedError: "",
		},
		{
			Name: "invalid_action_name",
			Task: PkgTask{
				TypeName:   "unknown type name",
				Path:       "somepath",
				NamedTask:  NamedTask{Name: "some name"},
				ActionType: 0,
			},
			ExpectedError: "unknown pkg task type: unknown type name",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Task.Validate()
			if tc.ExpectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.ExpectedError)
			}
		})
	}
}

func TestPkgTaskPath(t *testing.T) {
	task := PkgTask{
		Path: "somepath",
	}

	assert.Equal(t, "somepath", task.GetPath())
}

func TestPkgTaskName(t *testing.T) {
	task := PkgTask{
		TypeName: PkgRemoved,
	}

	assert.Equal(t, PkgRemoved, task.GetName())
}

func TestPkgTaskRequire(t *testing.T) {
	task := PkgTask{
		Require: []string{"require one", "require two"},
	}

	assert.Equal(t, []string{"require one", "require two"}, task.GetRequirements())
}

func TestPkgTaskString(t *testing.T) {
	task := PkgTask{
		Path:     "task1",
		TypeName: PkgUpgraded,
	}

	assert.Equal(t, fmt.Sprintf("task '%s' at path 'task1'", PkgUpgraded), task.String())
}

func TestPkgTaskExecution(t *testing.T) {
	testCases := []struct {
		Task               *PkgTask
		ExpectedResult     ExecutionResult
		RunnerMock         *appExec.SystemRunner
		PackageManagerMock *PackageManagerMock
		ExpectedCmdStrs    []string
		Name               string
	}{
		{
			Name: "execute one name install",
			Task: &PkgTask{
				ActionType: ActionInstall,
				TypeName:   PkgInstalled,
				Path:       "one name path",
				NamedTask:  NamedTask{Name: "vim"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "installation success",
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			PackageManagerMock: &PackageManagerMock{
				outputToGive: &PackageManagerExecutionResult{
					Output: "installation success",
				},
			},
		},
		{
			Name: "executing one onlyif condition with success",
			Task: &PkgTask{
				NamedTask: NamedTask{Name: "cmd lala"},
				OnlyIf:    []string{"check before lala"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "installation success",
				Comment: "some comment",
				Changes: map[string]string{
					"some change key": "some change value",
				},
			},
			ExpectedCmdStrs: []string{"check before lala"},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			PackageManagerMock: &PackageManagerMock{
				outputToGive: &PackageManagerExecutionResult{
					Output:  "installation success",
					Comment: "some comment",
					Changes: map[string]string{
						"some change key": "some change value",
					},
				},
			},
		},
		{
			Name: "executing one onlyif condition with skip execution",
			Task: &PkgTask{
				NamedTask: NamedTask{Name: "cmd with OnlyIf skipped"},
				OnlyIf:    []string{"check OnlyIf error"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					assert.Contains(t, cmd.String(), "check OnlyIf error")
					return appExec.RunError{Err: errors.New("some OnlyIfFailure")}
				},
			}},
			PackageManagerMock: &PackageManagerMock{},
		},
		{
			Name: "executing one unless condition with success",
			Task: &PkgTask{
				NamedTask: NamedTask{Name: "cmd stop"},
				Unless:    []string{"run unless stop"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"run unless stop"},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					cmdStr := cmd.String()
					if strings.Contains(cmdStr, "run unless stop") {
						return appExec.RunError{Err: errors.New("run unless stop failed")}
					}

					return nil
				},
			}},
			PackageManagerMock: &PackageManagerMock{},
		},
		{
			Name: "executing one unless condition with failure",
			Task: &PkgTask{
				NamedTask: NamedTask{Name: "cmd with unless failure"},
				Unless:    []string{"check unless failure"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			PackageManagerMock: &PackageManagerMock{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			executor := &PkgTaskExecutor{
				Runner:         tc.RunnerMock,
				PackageManager: tc.PackageManagerMock,
			}

			res := executor.Execute(context.Background(), tc.Task)
			assert.EqualValues(tt, tc.ExpectedResult.Err, res.Err)
			assert.EqualValues(tt, tc.ExpectedResult.IsSkipped, res.IsSkipped)
			assert.EqualValues(tt, tc.ExpectedResult.StdOut, res.StdOut)
			assert.EqualValues(tt, tc.ExpectedResult.StdErr, res.StdErr)
			assert.EqualValues(tt, tc.ExpectedResult.Comment, res.Comment)
			assert.EqualValues(tt, tc.ExpectedResult.Changes, res.Changes)

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
			assertPkgTaskEquals(tt, tc.Task, tc.PackageManagerMock.givenTask)
		})
	}
}

func TestInvalidTaskTypeExecution(t *testing.T) {
	runnerMock := &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
		Cmds: []*exec.Cmd{},
	}}
	pkgManager := &PackageManagerMock{}
	executor := &PkgTaskExecutor{
		Runner:         runnerMock,
		PackageManager: pkgManager,
	}

	res := executor.Execute(context.TODO(), &CmdRunTask{Path: "some path"})
	assert.Contains(t, res.Err.Error(), "to PkgTask")
}
