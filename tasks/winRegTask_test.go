//go:build windows
// +build windows

package tasks

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/winreg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWinRegTaskValidation(t *testing.T) {
	testCases := []struct {
		Name        string
		ExpectedErr string
		Task        WinRegTask
	}{
		{
			Name: "invalid_action_name",
			Task: WinRegTask{
				TypeName:   "unknown type name",
				Path:       "somepath",
				ActionType: 0,
			},
			ExpectedErr: "unknown win_reg task type: unknown type name",
		},
		{
			Name: "present, missing name",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:     "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", NameField),
		},
		{
			Name: "present, missing reg path",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:     "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", RegPathField),
		},
		{
			Name: "present, missing value",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				RegPath:    `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				// Val: "0",
				ValType: "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", ValField),
		},
		{
			Name: "present, missing type",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegPresent,
				Name:       "fDenyTSConnections",
				RegPath:    `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
				Val:        "0",
				// ValType:  "string",
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", ValTypeField),
		},
		{
			Name: "absent, missing name",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", NameField),
		},
		{
			Name: "absent, missing reg path",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				Name:       "fDenyTSConnections",
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", RegPathField),
		},
		{
			Name: "absent, missing name",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsent,
				// Name:       "fDenyTSConnections",
				RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", NameField),
		},
		{
			Name: "absent key, missing reg path",
			Task: WinRegTask{
				Path:       "winregpath",
				ActionType: ActionWinRegAbsentKey,
				// RegPath: `HKLM:\System\CurrentControlSet\Control\Terminal Server`,
			},
			ExpectedErr: fmt.Sprintf("empty required value at path 'winregpath.%s'", RegPathField),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Task.Validate()
			if tc.ExpectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.ExpectedErr)
			}
		})
	}
}

func TestWinRegTaskPath(t *testing.T) {
	task := WinRegTask{
		Path: "winregpath",
	}

	assert.Equal(t, "winregpath", task.GetPath())
}

func TestWinRegTaskName(t *testing.T) {
	task := WinRegTask{
		TypeName: WinRegPresent,
	}

	assert.Equal(t, WinRegPresent, task.GetTypeName())
}

func TestWinRegTaskString(t *testing.T) {
	task := WinRegTask{
		Path:     "task1",
		TypeName: WinRegAbsent,
	}

	assert.Equal(t, fmt.Sprintf("task '%s' at path 'task1'", WinRegAbsent), task.String())
}

func TestWinRegTaskRequire(t *testing.T) {
	task := WinRegTask{
		Require: []string{"require one", "require two"},
	}

	assert.Equal(t, []string{"require one", "require two"}, task.GetRequirements())
}

func TestShouldEnsureRegistryValueIsPresent(t *testing.T) {
	ctx := context.Background()

	executor := &WinRegTaskExecutor{}

	task := &WinRegTask{
		ActionType: ActionWinRegPresent,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winreg.RemoveValue(task.RegPath, task.Name)
	require.NoError(t, err)

	err = task.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, val, err := winreg.GetValue(task.RegPath, task.Name, "REG_SZ")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, task.Val, val)
}

func TestShouldEnsureRegistryValueIsAbsent(t *testing.T) {
	ctx := context.Background()

	executor := &WinRegTaskExecutor{}

	task := &WinRegTask{
		ActionType: ActionWinRegAbsent,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winreg.SetValue(task.RegPath, task.Name, task.Val, winreg.REG_SZ)
	require.NoError(t, err)

	err = task.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, _, err := winreg.GetValue(task.RegPath, task.Name, "REG_SZ")
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestShouldEnsureRegistryKeyIsAbsent(t *testing.T) {
	ctx := context.Background()

	executor := &WinRegTaskExecutor{}

	task := &WinRegTask{
		ActionType: ActionWinRegAbsentKey,
		Path:       "set-value-1",
		Name:       "testValue",
		RegPath:    `HKLM:\Software\TacoScript\UnitTestRun`,
		Val:        "new value",
		ValType:    "REG_SZ",
	}

	_, _, err := winreg.SetValue(task.RegPath, task.Name, task.Val, winreg.REG_SZ)
	require.NoError(t, err)

	err = task.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)

	assert.True(t, task.Updated)
	assert.False(t, res.IsSkipped)

	found, _, err := winreg.GetValue(task.RegPath, task.Name, winreg.REG_SZ)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestShouldEnsureConditionalsAreChecked(t *testing.T) {
	t.Skip()
}