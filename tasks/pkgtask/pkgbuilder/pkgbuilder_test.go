package pkgbuilder

import (
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/namedtask"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []interface{}
		expectedTask  *pkgtask.PTask
		expectedError string
	}{
		{
			typeName: pkgtask.TaskTypePkgInstalled,
			path:     "vim",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "vim"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "cmd.exe"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.Version, Value: "1.0.1"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.Refresh, Value: 1}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
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
			expectedTask: &pkgtask.PTask{
				ActionType:    pkgtask.ActionInstall,
				TypeName:      pkgtask.TaskTypePkgInstalled,
				Path:          "vim",
				Named:         namedtask.NamedTask{Name: "vim"},
				Shell:         "cmd.exe",
				Version:       "1.0.1",
				ShouldRefresh: true,
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
			typeName: pkgtask.TaskTypePkgUpgraded,
			path:     "git",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "git"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.Version, Value: "2.0.2"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.Refresh, Value: "false"}},
			},
			expectedTask: &pkgtask.PTask{
				ActionType:    pkgtask.ActionUpdate,
				TypeName:      pkgtask.TaskTypePkgUpgraded,
				Path:          "git",
				Named:         namedtask.NamedTask{Name: "git"},
				Version:       "2.0.2",
				ShouldRefresh: false,
			},
		},
		{
			typeName: pkgtask.TaskTypePkgRemoved,
			path:     "nano",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NamesField, Value: []interface{}{
					"nano",
					"git",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.Refresh, Value: ""}},
				yaml.MapSlice{yaml.MapItem{Key: "someField", Value: "someValue"}},
			},
			expectedTask: &pkgtask.PTask{
				ActionType: pkgtask.ActionUninstall,
				TypeName:   pkgtask.TaskTypePkgRemoved,
				Path:       "nano",
				Named: namedtask.NamedTask{Names: []string{
					"nano",
					"git",
				}},
				ShouldRefresh: false,
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

			actualTask, ok := actualTaskI.(*pkgtask.PTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assertPkgTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertPkgTaskEquals(t *testing.T, expectedTask, actualTask *pkgtask.PTask) {
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
