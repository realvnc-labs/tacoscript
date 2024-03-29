package pkgtask

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"testing"

	appExec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
	"github.com/realvnc-labs/tacoscript/tasks/shared/executionresult"
	"github.com/realvnc-labs/tacoscript/tasks/shared/names"
	"github.com/stretchr/testify/assert"
)

type PackageManagerMock struct {
	givenCtx     context.Context
	givenTask    *Task
	outputToGive *ExecutionResult
	errToGive    error
}

func (pmm *PackageManagerMock) ExecuteTask(ctx context.Context, t *Task) (res *ExecutionResult, err error) {
	pmm.givenCtx = ctx
	pmm.givenTask = t

	return pmm.outputToGive, pmm.errToGive
}

func TestPkgTaskValidation(t *testing.T) {
	testCases := []struct {
		Name          string
		ExpectedError string
		InputTask     Task
	}{
		{
			Name: "missing_name_and_names",
			InputTask: Task{
				Path:       "somepath",
				ActionType: ActionUpdate,
			},
			ExpectedError: fmt.Sprintf(
				"empty required value at path 'somepath.%s', empty required values at path 'somepath.%s'",
				tasks.NameField,
				tasks.NamesField,
			),
		},
		{
			Name: "valid_task_name",
			InputTask: Task{
				Named:      names.TaskNames{Name: "some name"},
				ActionType: ActionUninstall,
			},
			ExpectedError: "",
		},
		{
			Name: "valid_task_names",
			InputTask: Task{
				Named:      names.TaskNames{Names: []string{"some name1", "some name 2"}},
				ActionType: ActionInstall,
			},
			ExpectedError: "",
		},
		{
			Name: "invalid_action_name",
			InputTask: Task{
				TypeName:   "unknown type name",
				Path:       "somepath",
				Named:      names.TaskNames{Name: "some name"},
				ActionType: 0,
			},
			ExpectedError: "unknown pkg task type: unknown type name",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.InputTask.Validate(runtime.GOOS)
			if tc.ExpectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.ExpectedError)
			}
		})
	}
}

func TestPkgTaskPath(t *testing.T) {
	task := Task{
		Path: "somepath",
	}

	assert.Equal(t, "somepath", task.GetPath())
}

func TestPkgTaskName(t *testing.T) {
	task := Task{
		TypeName: TaskTypePkgRemoved,
	}

	assert.Equal(t, TaskTypePkgRemoved, task.GetTypeName())
}

func TestPkgTaskRequire(t *testing.T) {
	task := Task{
		Require: []string{"require one", "require two"},
	}

	assert.Equal(t, []string{"require one", "require two"}, task.GetRequirements())
}

func TestPkgTaskString(t *testing.T) {
	task := Task{
		Path:     "task1",
		TypeName: TaskTypePkgUpgraded,
	}

	assert.Equal(t, fmt.Sprintf("task '%s' at path 'task1'", TaskTypePkgUpgraded), task.String())
}

func TestPkgTaskExecution(t *testing.T) {
	testCases := []struct {
		InputTask          *Task
		ExpectedResult     executionresult.ExecutionResult
		RunnerMock         *appExec.SystemRunner
		PackageManagerMock *PackageManagerMock
		ExpectedCmdStrs    []string
		Name               string
	}{
		{
			Name: "execute one name install",
			InputTask: &Task{
				ActionType: ActionInstall,
				TypeName:   TaskTypePkgInstalled,
				Path:       "one name path",
				Named:      names.TaskNames{Name: "vim"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: false,
				Err:       nil,
				StdOut:    "installation success",
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			PackageManagerMock: &PackageManagerMock{
				outputToGive: &ExecutionResult{
					Output: "installation success",
				},
			},
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
				StdOut:    "installation success",
				Comment:   "some comment",
				Changes: map[string]string{
					"some change key": "some change value",
				},
			},
			ExpectedCmdStrs: []string{"check before lala"},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			PackageManagerMock: &PackageManagerMock{
				outputToGive: &ExecutionResult{
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
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with OnlyIf skipped"},
				OnlyIf: []string{"check OnlyIf error"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: true,
			},
			ExpectedCmdStrs: []string{},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					return appExec.RunError{Err: errors.New("some OnlyIfFailure")}
				},
			}},
			PackageManagerMock: &PackageManagerMock{},
		},
		{
			Name: "executing one unless condition with success",
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd stop"},
				Unless: []string{"run unless stop"},
			},
			ExpectedResult: executionresult.ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStrs: []string{"run unless stop"},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
				Callback: func(cmd *exec.Cmd) error {
					return appExec.RunError{Err: errors.New("run unless stop failed")}
				},
			}},
			PackageManagerMock: &PackageManagerMock{},
		},
		{
			Name: "executing one unless condition with failure",
			InputTask: &Task{
				Named:  names.TaskNames{Name: "cmd with unless failure"},
				Unless: []string{"check unless failure"},
			},
			ExpectedResult: executionresult.ExecutionResult{
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
			executor := &Executor{
				Runner:         tc.RunnerMock,
				PackageManager: tc.PackageManagerMock,
			}

			res := executor.Execute(context.Background(), tc.InputTask)
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
			if tc.ExpectedResult.IsSkipped {
				return
			}

			assert.Equal(tt, len(tc.ExpectedCmdStrs), len(cmds))
			if tc.PackageManagerMock.givenTask != nil {
				assertPkgTaskEquals(tt, tc.InputTask, tc.PackageManagerMock.givenTask)
			}
		})
	}
}

func TestInvalidTaskTypeExecution(t *testing.T) {
	runnerMock := &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
		Cmds: []*exec.Cmd{},
	}}
	pkgManager := &PackageManagerMock{}
	executor := &Executor{
		Runner:         runnerMock,
		PackageManager: pkgManager,
	}

	res := executor.Execute(context.TODO(), &cmdrun.Task{Path: "some path"})
	assert.Contains(t, res.Err.Error(), "to Task")
}

func assertPkgTaskEquals(t *testing.T, expectedTask, actualTask *Task) {
	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName)
	assert.Equal(t, expectedTask.Path, actualTask.Path)
	assert.Equal(t, expectedTask.Named.Name, actualTask.Named.Name)
	assert.Equal(t, expectedTask.Named.Names, actualTask.Named.Names)
	assert.Equal(t, expectedTask.Require, actualTask.Require)
	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
	assert.Equal(t, expectedTask.ActionType, actualTask.ActionType)
	assert.Equal(t, expectedTask.ShouldRefresh, actualTask.ShouldRefresh)
	assert.Equal(t, expectedTask.Version, actualTask.Version)
	assert.Equal(t, expectedTask.Shell, actualTask.Shell)
}
