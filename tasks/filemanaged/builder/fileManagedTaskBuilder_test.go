package builder

import (
	"net/url"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	"github.com/realvnc-labs/tacoscript/utils"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestFileManagedTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *filemanaged.FileManagedTask
		expectedError string
	}{
		{
			typeName: "fileManagedType",
			path:     "fileManagedPath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "C:\temp\npp.7.8.8.Installer.x64.exe"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SourceField,
					Value: "https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe",
				}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SourceHashField, Value: "79eef25f9b0b2c642c62b7f737d4f53f"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.MakeDirsField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ReplaceField, Value: false}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipVerifyField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "C:\\Program Files\notepad++\notepad++.exe"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
			},
			expectedTask: &filemanaged.FileManagedTask{
				TypeName: "fileManagedType",
				Path:     "fileManagedPath",
				Name:     "C:\temp\npp.7.8.8.Installer.x64.exe",
				Source: utils.Location{
					IsURL: true,
					URL: &url.URL{
						Scheme: "https",
						Host:   "github.com",
						Path:   "/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe",
					},
					RawLocation: "https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe",
				},
				SourceHash: "79eef25f9b0b2c642c62b7f737d4f53f",
				MakeDirs:   true,
				Replace:    false,
				SkipVerify: true,
				Creates:    []string{"C:\\Program Files\notepad++\notepad++.exe"},
				Shell:      "someshell",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			taskBuilder := FileManagedTaskBuilder{}
			actualTaskI, err := taskBuilder.Build(
				tc.typeName,
				tc.path,
				tc.values,
			)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			if err != nil {
				return
			}

			actualTask, ok := actualTaskI.(*filemanaged.FileManagedTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assertFileManagedTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertFileManagedTaskEquals(t *testing.T, expectedTask, actualTask *filemanaged.FileManagedTask) {
	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName)
	assert.Equal(t, expectedTask.User, actualTask.User)
	assert.Equal(t, expectedTask.Group, actualTask.Group)
	assert.Equal(t, expectedTask.Path, actualTask.Path)
	assert.Equal(t, expectedTask.SkipVerify, actualTask.SkipVerify)
	assert.Equal(t, expectedTask.Name, actualTask.Name)
	assert.Equal(t, expectedTask.Mode.String(), actualTask.Mode.String())
	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName)
	assert.Equal(t, expectedTask.Source, actualTask.Source)
	assert.Equal(t, expectedTask.SourceHash, actualTask.SourceHash)
	assert.Equal(t, expectedTask.Replace, actualTask.Replace)
	assert.Equal(t, expectedTask.MakeDirs, actualTask.MakeDirs)
	assert.Equal(t, expectedTask.Encoding, actualTask.Encoding)
	assert.Equal(t, expectedTask.Contents, actualTask.Contents)
	assert.Equal(t, expectedTask.Require, actualTask.Require)
	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
}
