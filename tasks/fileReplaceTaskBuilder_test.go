package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestFileReplaceTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *FileReplaceTask
		expectedError string
	}{
		{
			typeName: "fileReplaceType",
			path:     "fileReplacePath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: NameField, Value: "/tmp/file-to-replace.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: PatternField, Value: "username = jack"}},
				yaml.MapSlice{yaml.MapItem{Key: ReplField, Value: "username = jill"}},
				yaml.MapSlice{yaml.MapItem{Key: CountField, Value: "10"}},
				yaml.MapSlice{yaml.MapItem{Key: AppendIfNotFoundField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: PrependIfNotFoundField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: NotFoundContentField, Value: "new text when not found"}},
				yaml.MapSlice{yaml.MapItem{Key: BackupExtensionField, Value: "bak"}},
				yaml.MapSlice{yaml.MapItem{Key: MaxFileSizeField, Value: "100k"}},

				yaml.MapSlice{yaml.MapItem{Key: CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: "/tmp/unless-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: ShellField, Value: "someshell"}},
			},
			expectedTask: &FileReplaceTask{
				TypeName:          "fileReplaceType",
				Path:              "fileReplacePath",
				Name:              "/tmp/file-to-replace.txt",
				Pattern:           "username = jack",
				Repl:              "username = jill",
				Count:             10,
				AppendIfNotFound:  true,
				PrependIfNotFound: true,
				NotFoundContent:   "new text when not found",
				BackupExtension:   "bak",
				MaxFileSize:       "100k",

				Creates: []string{"/tmp/creates-file.txt"},
				OnlyIf:  []string{"/tmp/onlyif-file.txt"},
				Unless:  []string{"/tmp/unless-file.txt"},
				Require: []string{"/tmp/required-file.txt"},

				Shell: "someshell",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			taskBuilder := FileReplaceTaskBuilder{}
			task, err := taskBuilder.Build(
				tc.typeName,
				tc.path,
				tc.values,
			)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, tc.expectedError)
				return
			}
			require.NoError(t, err)

			actualTask, ok := task.(*FileReplaceTask)
			require.True(t, ok)

			assertFileReplaceTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertFileReplaceTaskEquals(t *testing.T, expectedTask, actualTask *FileReplaceTask) {
	assert.Equal(t, expectedTask.Name, actualTask.Name)
	assert.Equal(t, expectedTask.Path, actualTask.GetPath())

	assert.Equal(t, expectedTask.Pattern, actualTask.Pattern)
	assert.Equal(t, expectedTask.Repl, actualTask.Repl)
	assert.Equal(t, expectedTask.Count, actualTask.Count)
	assert.Equal(t, expectedTask.AppendIfNotFound, actualTask.AppendIfNotFound)
	assert.Equal(t, expectedTask.PrependIfNotFound, actualTask.PrependIfNotFound)
	assert.Equal(t, expectedTask.NotFoundContent, actualTask.NotFoundContent)
	assert.Equal(t, expectedTask.BackupExtension, actualTask.BackupExtension)
	assert.Equal(t, expectedTask.MaxFileSize, actualTask.MaxFileSize)

	assert.Equal(t, expectedTask.Require, actualTask.Require)

	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
}
