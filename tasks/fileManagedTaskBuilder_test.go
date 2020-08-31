package tasks

import (
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/utils"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileManagedTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		ctx           []map[string]interface{}
		expectedTask  *FileManagedTask
		expectedError string
	}{
		{
			typeName: "fileManagedType",
			path:     "fileManagedPath",
			ctx: []map[string]interface{}{
				{
					NameField:       "C:\temp\npp.7.8.8.Installer.x64.exe",
					SourceField:     "https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe",
					SourceHashField: "79eef25f9b0b2c642c62b7f737d4f53f",
					MakeDirsField:   true,
					ReplaceField:    false,
					SkipVerifyField: true,
					CreatesField:    "C:\\Program Files\notepad++\notepad++.exe",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "fileManagedType",
				Path:     "fileManagedPath",
				Name:     "C:\temp\npp.7.8.8.Installer.x64.exe",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
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
			},
		},
		{
			typeName: "fileManagedType2",
			path:     "fileManagedPath2",
			ctx: []map[string]interface{}{
				{
					NameField: "/tmp/my-file.txt",
					ContentsField: `My file content
goes here
Funny file`,
					UserField:     "root",
					GroupField:    "www-data",
					ModeField:     0755,
					EncodingField: "UTF-8",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "fileManagedType2",
				Path:     "fileManagedPath2",
				Name:     "/tmp/my-file.txt",
				Contents: `My file content
goes here
Funny file`,
				User:     "root",
				Group:    "www-data",
				Mode:     0755,
				Encoding: "UTF-8",
			},
		},
		{
			typeName: "manyCreatesType",
			path:     "manyCreatesPath",
			ctx: []map[string]interface{}{
				{
					NameField: "many creates command",
					CreatesField: []interface{}{
						"create one",
						"create two",
						"create three",
					},
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
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "manyCreatesType",
				Path:     "manyCreatesPath",
				Name:     "many creates command",
				Creates: []string{
					"create one",
					"create two",
					"create three",
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
			typeName: "localFileSource1",
			path:     "localFileSource1Path",
			ctx: []map[string]interface{}{
				{
					SourceField: "file:///someFile/ru",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "localFileSource1",
				Path:     "localFileSource1Path",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "/someFile/ru",
					RawLocation: "file:///someFile/ru",
				},
			},
		},
		{
			typeName: "localFileSource3",
			path:     "localFileSource3Path",
			ctx: []map[string]interface{}{
				{
					SourceField: "/Users/space.txt",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "localFileSource3",
				Path:     "localFileSource3Path",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   `/Users/space.txt`,
					RawLocation: "/Users/space.txt",
				},
			},
		},
		{
			typeName: "localFileSource4",
			path:     "localFileSource4Path",
			ctx: []map[string]interface{}{
				{
					SourceField: "last.txt",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "localFileSource4",
				Path:     "localFileSource4Path",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   `last.txt`,
					RawLocation: "last.txt",
				},
			},
		},
		{
			typeName: "http(s)url",
			path:     "http(s)urlPath",
			ctx: []map[string]interface{}{
				{
					SourceField: "//github.com/some/path",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "http(s)url",
				Path:     "http(s)urlPath",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
						Host: "github.com",
						Path: "/some/path",
					},
					RawLocation: "//github.com/some/path",
				},
			},
		},
		{
			typeName: "invalid_filemode",
			path:     "invalid_filemode_path",
			ctx: []map[string]interface{}{
				{
					ModeField: "dfasdf",
				},
			},
			expectedError: fmt.Sprintf("invalid file mode value 'dfasdf' at path 'invalid_filemode_path.%s'", ModeField),
		},
		{
			typeName: "correct_string_mode",
			path:     "correct_string_mode_path",
			ctx: []map[string]interface{}{
				{
					NameField: "correct_string_mode.txt",
					ModeField: "0777",
				},
			},
			expectedTask: &FileManagedTask{
				TypeName: "correct_string_mode",
				Path:     "correct_string_mode_path",
				Mode: os.FileMode(0777),
				Name: "correct_string_mode.txt",
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

			actualTask, ok := actualTaskI.(*FileManagedTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assertFileManagedTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertFileManagedTaskEquals(t *testing.T, expectedTask, actualTask *FileManagedTask) {
	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName)
	assert.Equal(t, expectedTask.User, actualTask.User)
	assert.Equal(t, expectedTask.Group, actualTask.Group)
	assert.Equal(t, expectedTask.Path, actualTask.Path)
	assert.Equal(t, expectedTask.SkipVerify, actualTask.SkipVerify)
	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.Name, actualTask.Name)
	assert.Equal(t, expectedTask.Mode.String(), actualTask.Mode.String())
	assert.Equal(t, expectedTask.TypeName, actualTask.TypeName)
	assert.Equal(t, expectedTask.Source, actualTask.Source)
	assert.Equal(t, expectedTask.SourceHash, actualTask.SourceHash)
	assert.Equal(t, expectedTask.Replace, actualTask.Replace)
	assert.Equal(t, expectedTask.MakeDirs, actualTask.MakeDirs)
	assert.Equal(t, expectedTask.Require, actualTask.Require)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Encoding, actualTask.Encoding)
	assert.Equal(t, expectedTask.Contents, actualTask.Contents)
}