package builder

import (
	"testing"

	tasks "github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filereplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestFileReplaceTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *filereplace.FrTask
		expectedError string
	}{
		{
			typeName: "fileReplaceType",
			path:     "fileReplacePath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.NameField, Value: "/tmp/file-to-replace.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.PatternField, Value: "username = jack"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ReplField, Value: "username = jill"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CountField, Value: "10"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.AppendIfNotFoundField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.PrependIfNotFoundField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.NotFoundContentField, Value: "new text when not found"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BackupExtensionField, Value: "bak"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.MaxFileSizeField, Value: "100k"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: "/tmp/unless-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
			},
			expectedTask: &filereplace.FrTask{
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

			actualTask, ok := task.(*filereplace.FrTask)
			require.True(t, ok)

			assertFileReplaceTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertFileReplaceTaskEquals(t *testing.T, expectedTask, actualTask *filereplace.FrTask) {
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
