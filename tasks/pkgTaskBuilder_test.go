package tasks

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPkgTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []map[string]interface{}
		expectedTask  *PkgTask
		expectedError string
	}{
		{
			typeName: PkgInstalled,
			path:     "vim",
			ctx: []map[string]interface{}{
				{
					NameField:  "vim",
					ShellField: "cmd.exe",
					Version:    "1.0.1",
					Refresh:    1,
					RequireField: []interface{}{
						"req one",
						"req two",
						"req three",
					},
					OnlyIf: []interface{}{
						"OnlyIf one",
						"OnlyIf two",
						"OnlyIf three",
					},
					Unless: []interface{}{
						"Unless one",
					},
				},
			},
			expectedTask: &PkgTask{
				ActionType:    ActionInstall,
				TypeName:      PkgInstalled,
				Path:          "vim",
				NamedTask:     NamedTask{Name: "vim"},
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
			ctx: []map[string]interface{}{
				{
					NameField: "git",
					Version:   "2.0.2",
					Refresh:   "false",
				},
			},
			expectedTask: &PkgTask{
				ActionType:    ActionUpdate,
				TypeName:      PkgUpgraded,
				Path:          "git",
				NamedTask:     NamedTask{Name: "git"},
				Version:       "2.0.2",
				ShouldRefresh: false,
			},
		},
		{
			typeName: PkgRemoved,
			path:     "nano",
			ctx: []map[string]interface{}{
				{
					NamesField: []interface{}{
						"nano",
						"git",
					},
					Refresh:     "",
					"someField": "someValue",
				},
			},
			expectedTask: &PkgTask{
				ActionType: ActionUninstall,
				TypeName:   PkgRemoved,
				Path:       "nano",
				NamedTask: NamedTask{Names: []string{
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
	assert.Equal(t, expectedTask.Name, actualTask.Name)
	assert.Equal(t, expectedTask.Names, actualTask.Names)
	assert.Equal(t, expectedTask.Require, actualTask.Require)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
	assert.Equal(t, expectedTask.ActionType, actualTask.ActionType)
	assert.Equal(t, expectedTask.ShouldRefresh, actualTask.ShouldRefresh)
	assert.Equal(t, expectedTask.Version, actualTask.Version)
	assert.Equal(t, expectedTask.Shell, actualTask.Shell)
}
