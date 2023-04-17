//go:build windows
// +build windows

package winreg

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/support/winregistry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWinRegTaskValidation(t *testing.T) {
	testCases := []struct {
		Name        string
		ExpectedErr string
		InputTask   Task
	}{
		{
			Name: "invalid_action_name",
			InputTask: Task{
				TypeName:   "unknown type name",
				Path:       "somepath",
				ActionType: 0,
			},
			ExpectedErr: "unknown win_reg task type: unknown type name",
		},
		{
			Name: "present, missing name",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:     "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.NameField),
		},
		{
			Name: "present, missing reg path",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:     "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.RegPathField),
		},
		{
			Name: "present, missing value",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				RegPath:    `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				// Val: "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.ValField),
		},
		{
			Name: "present, missing type",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				RegPath:    `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:        "0",
				// ValType:  "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.ValTypeField),
		},
		{
			Name: "absent, missing name",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.NameField),
		},
		{
			Name: "absent, missing reg path",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				Name:       "fDenyTSConnections",
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.RegPathField),
		},
		{
			Name: "absent, missing name",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.NameField),
		},
		{
			Name: "absent key, missing reg path",
			InputTask: Task{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsentKey,
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", tasks.RegPathField),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.InputTask.Validate(runtime.GOOS)
			if tc.ExpectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.ExpectedErr)
			}
		})
	}
}

func TestWinRegTaskPath(t *testing.T) {
	task := Task{
		Path: "winregpath",
	}

	assert.Equal(t, "winregpath", task.GetPath())
}

func TestWinRegTaskName(t *testing.T) {
	task := Task{
		TypeName: TaskTypeWinRegPresent,
	}

	assert.Equal(t, TaskTypeWinRegPresent, task.GetTypeName())
}

func TestWinRegTaskString(t *testing.T) {
	task := Task{
		Path:     "task1",
		TypeName: TaskTypeWinRegAbsent,
	}

	assert.Equal(t, fmt.Sprintf("task '%s' at path 'task1'", TaskTypeWinRegAbsent), task.String())
}

func TestWinRegTaskRequire(t *testing.T) {
	task := Task{
		Require: []string{"require one", "require two"},
	}

	assert.Equal(t, []string{"require one", "require two"}, task.GetRequirements())
}

func TestShouldEnsureRegistryValueIsPresent(t *testing.T) {
	ctx := context.Background()

	executor := &Executor{}

	task := &Task{
		ActionType: ActionWinRegPresent,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winregistry.RemoveValue(task.RegPath, task.Name)
	require.NoError(t, err)

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, val, err := winregistry.GetValue(task.RegPath, task.Name, "REG_SZ")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, task.Val, val)
}

func TestShouldEnsureRegistryValueIsAbsent(t *testing.T) {
	ctx := context.Background()

	executor := &Executor{}

	task := &Task{
		ActionType: ActionWinRegAbsent,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winregistry.SetValue(task.RegPath, task.Name, task.Val, winregistry.REG_SZ)
	require.NoError(t, err)

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, _, err := winregistry.GetValue(task.RegPath, task.Name, "REG_SZ")
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestShouldEnsureRegistryKeyIsAbsent(t *testing.T) {
	ctx := context.Background()

	executor := &Executor{}

	task := &Task{
		ActionType: ActionWinRegAbsentKey,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winregistry.SetValue(task.RegPath, task.Name, task.Val, winregistry.REG_SZ)
	require.NoError(t, err)

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, _, err := winregistry.GetValue(task.RegPath, task.Name, winregistry.REG_SZ)
	assert.NoError(t, err)
	assert.False(t, found)
}
