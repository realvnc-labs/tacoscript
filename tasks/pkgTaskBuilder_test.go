package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestPkgTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []interface{}
		expectedTask  *PkgTask
		expectedError string
	}{
		{
			typeName: PkgInstalled,
			path:     "vim",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "vim"}},
				yaml.MapSlice{yaml.MapItem{Key: ShellField, Value: "cmd.exe"}},
				yaml.MapSlice{yaml.MapItem{Key: Version, Value: "1.0.1"}},
				yaml.MapSlice{yaml.MapItem{Key: Refresh, Value: 1}},
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: []interface{}{
					"req one",
					"req two",
					"req three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: []interface{}{
					"Unless one",
				}}},
			},
			expectedTask: &PkgTask{
				ActionType:    ActionInstall,
				TypeName:      PkgInstalled,
				Path:          "vim",
				Named:         NamedTask{Name: "vim"},
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
			typeName: PkgUpgraded,
			path:     "git",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "git"}},
				yaml.MapSlice{yaml.MapItem{Key: Version, Value: "2.0.2"}},
				yaml.MapSlice{yaml.MapItem{Key: Refresh, Value: "false"}},
			},
			expectedTask: &PkgTask{
				ActionType:    ActionUpdate,
				TypeName:      PkgUpgraded,
				Path:          "git",
				Named:         NamedTask{Name: "git"},
				Version:       "2.0.2",
				ShouldRefresh: false,
			},
		},
		{
			typeName: PkgRemoved,
			path:     "nano",
			ctx: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NamesField, Value: []interface{}{
					"nano",
					"git",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: Refresh, Value: ""}},
				yaml.MapSlice{yaml.MapItem{Key: "someField", Value: "someValue"}},
			},
			expectedTask: &PkgTask{
				ActionType: ActionUninstall,
				TypeName:   PkgRemoved,
				Path:       "nano",
				Named: NamedTask{Names: []string{
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
			taskBuilder := PkgTaskBuilder{}
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

			actualTask, ok := actualTaskI.(*PkgTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assertPkgTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertPkgTaskEquals(t *testing.T, expectedTask, actualTask *PkgTask) {
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
