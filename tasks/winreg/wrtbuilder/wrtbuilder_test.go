package wrtbuilder

import (
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/winreg"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []interface{}
		expectedTask  *winreg.Task
		expectedError string
	}{
		{
			typeName: winreg.TaskTypeWinRegPresent,
			path:     "WinRegPath",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "fDenyTSConnections"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RegPathField, Value: "HKLM:\\System\\CurrentControlSet\\Control\\Terminal Server"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ValField, Value: "0"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ValTypeField, Value: "REG_SZ"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
					"creates one",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"Unless one",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
			},
			expectedTask: &winreg.Task{
				ActionType: winreg.ActionWinRegPresent,
				TypeName:   winreg.TaskTypeWinRegPresent,
				Path:       "WinRegPath",
				Name:       "fDenyTSConnections",
				RegPath:    "HKLM:\\System\\CurrentControlSet\\Control\\Terminal Server",
				Val:        "0",
				ValType:    "REG_SZ",
				Require: []string{
					"req one",
					"req two",
					"req three",
				},
				Creates: []string{
					"creates one",
				},
				Unless: []string{
					"Unless one",
				},
				OnlyIf: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
				Shell: "someshell",
			},
		},
		{
			typeName: winreg.TaskTypeWinRegAbsent,
			path:     "",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "VMware User Process"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RegPathField, Value: `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
					"creates one",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"Unless one",
				}}},
			},
			expectedTask: &winreg.Task{
				ActionType: winreg.ActionWinRegAbsent,
				TypeName:   winreg.TaskTypeWinRegAbsent,
				Name:       "VMware User Process",
				RegPath:    `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
				Creates: []string{
					"creates one",
				},
				Unless: []string{
					"Unless one",
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
			typeName: winreg.TaskTypeWinRegAbsentKey,
			path:     "",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "VMware User Process"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RegPathField, Value: `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: []interface{}{
					"creates one",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"Unless one",
				}}},
			},
			expectedTask: &winreg.Task{
				ActionType: winreg.ActionWinRegAbsentKey,
				TypeName:   winreg.TaskTypeWinRegAbsentKey,
				Name:       "VMware User Process",
				RegPath:    `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
				Creates: []string{
					"creates one",
				},
				Unless: []string{
					"Unless one",
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
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			taskBuilder := TaskBuilder{}
			actualTaskI, err := taskBuilder.Build(
				tc.typeName,
				tc.path,
				tc.ctx,
			)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			if err != nil {
				return
			}

			actualTask, ok := actualTaskI.(*winreg.Task)
			assert.True(t, ok)
			if !ok {
				return
			}

			assertWinRegTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertWinRegTaskEquals(t *testing.T, expectedTask, actualTask *winreg.Task) {
	t.Helper()

	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName, "TypeName")
	assert.Equal(t, expectedTask.Path, actualTask.Path, "Path")

	assert.Equal(t, expectedTask.Require, actualTask.Require, "Require")
	assert.Equal(t, expectedTask.Creates, actualTask.Creates, "Creates")
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf, "OnlyIf")
	assert.Equal(t, expectedTask.Unless, actualTask.Unless, "Unless")

	assert.Equal(t, expectedTask.ActionType, actualTask.ActionType, "ActionType")
	assert.Equal(t, expectedTask.RegPath, actualTask.RegPath, "RegPath")
	assert.Equal(t, expectedTask.Name, actualTask.Name, "Name")

	if expectedTask.ActionType == winreg.ActionWinRegPresent {
		assert.Equal(t, expectedTask.Val, actualTask.Val, "Val")
		assert.Equal(t, expectedTask.ValType, actualTask.ValType, "ValType")
	}
}
