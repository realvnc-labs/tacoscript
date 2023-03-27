package filereplace

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReplaceTaskValidation(t *testing.T) {
	testCases := []struct {
		Name             string
		Task             FileReplaceTask
		ExpectedErrorStr string
	}{
		{
			Name: "missing_name",
			Task: FileReplaceTask{
				Path:    "somepath",
				Pattern: "search for this text",
			},
			ExpectedErrorStr: fmt.Errorf("empty required value at path 'somepath.%s'", tasks.NameField).Error(),
		},
		{
			Name: "valid task",
			Task: FileReplaceTask{
				Name:    "some p",
				Pattern: "search for this text",
			},
		},
		{
			Name: "bad pattern",
			Task: FileReplaceTask{
				Name:    "some p",
				Pattern: "*.txt",
			},
			ExpectedErrorStr: "error parsing regexp",
		},
		{
			Name: "invalid file size units",
			Task: FileReplaceTask{
				Name:        "some p",
				Pattern:     "search for this text",
				MaxFileSize: "100c",
			},
			ExpectedErrorStr: conv.ErrFileSizeInvalidUnits.Error(),
		},
		{
			Name: "append and prepend",
			Task: FileReplaceTask{
				Name:              "some p",
				Pattern:           "search for this text",
				AppendIfNotFound:  true,
				PrependIfNotFound: true,
			},
			ExpectedErrorStr: ErrAppendAndPrependSetAtTheSameTime.Error(),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Task.Validate(runtime.GOOS)

			if tc.ExpectedErrorStr != "" {
				assert.ErrorContains(t, err, tc.ExpectedErrorStr)
				return
			}

			assert.NoError(t, err)
			// if we've a valid task then check the default for max_file_size
			if strings.Contains(tc.Name, "valid task") {
				assert.Equal(t, defaultMaxFileSize, tc.Task.MaxFileSize)
				fileSize, err := conv.ConvertToFileSize(defaultMaxFileSize)
				require.NoError(t, err)
				assert.Equal(t, fileSize, tc.Task.maxFileSizeCalculated)
			}
		})
	}
}

var (
	simpleTestFileContents = `
this is a test file
`

	simpleTestFileContentWithRepetition = `
line 1
line 2
line 3
line 4
`
)

func getTestFilename() (name string) {
	name = os.TempDir() + "/testfile.txt"
	return name
}

func WriteTestFile(t *testing.T, filename string, contents string) {
	t.Helper()

	err := os.WriteFile(filename, []byte(contents), 0600)
	require.NoError(t, err)
}

func ReadFileContents(t *testing.T, filename string) (contents string) {
	t.Helper()

	bytes, err := os.ReadFile(filename)
	require.NoError(t, err)

	return string(bytes)
}

func TestShouldFailWhenTargetFileNotFound(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "a test",
		Repl:    "not a test",
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)

	var expectedErrStr string
	if runtime.GOOS == "windows" {
		expectedErrStr = fmt.Sprintf("CreateFile %s: The system cannot find the file specified.", testFilename)
	} else {
		expectedErrStr = fmt.Sprintf("stat %s: no such file or directory", testFilename)
	}

	require.EqualError(t, res.Err, expectedErrStr)
}

func TestShouldFailWhenMandatoryParamsMissing(t *testing.T) {
	cases := []struct {
		name           string
		task           FileReplaceTask
		expectedErrStr string
	}{
		{
			name: "all required params",
			task: FileReplaceTask{
				Name:    "test",
				Pattern: "pattern text",
			},
		},
		{
			name: "missing name",
			task: FileReplaceTask{
				// Name:    "test",
				Pattern: "pattern text",
			},
			expectedErrStr: "empty required value at path '.name'",
		},
		{
			name: "missing pattern",
			task: FileReplaceTask{
				Name: "test",
				// Pattern: "pattern text",
			},
			expectedErrStr: "empty required value at path '.pattern'",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate(runtime.GOOS)
			if tc.expectedErrStr != "" {
				assert.Contains(t, err.Error(), tc.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShouldMakeBackupOfOriginalFileWhenBackupExtensionSet(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContents)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:            "replace-1",
		Name:            testFilename,
		Pattern:         "a test",
		Repl:            "not a test",
		BackupExtension: "bak",
	}
	defer os.Remove(testFilename + "." + task.BackupExtension)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.FileExists(t, testFilename+"."+task.BackupExtension)
}

func TestShouldNotMakeBackupOfOriginalFileWhenBackupExtensionNotSet(t *testing.T) {
	ctx := context.Background()

	testConfigFilename := getTestFilename()

	WriteTestFile(t, testConfigFilename, simpleTestFileContents)
	defer os.Remove(testConfigFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testConfigFilename,
		Pattern: "a test",
		Repl:    "not a test",
		// BackupExtension: "bak",
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.NoFileExists(t, utils.GetBackupFilename(testConfigFilename, "bak"))
}

func TestShouldReplaceAllMatchingItems(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "line",
		Repl:    "new line",
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "File updated")
	assert.Equal(t, res.Changes["count"], "4 replacement(s) made")

	contents := ReadFileContents(t, testFilename)

	matchingCount := strings.Count(contents, "new line")
	assert.Equal(t, 4, matchingCount)
}

func TestShouldNotReplaceAnything(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	origContents := ReadFileContents(t, testFilename)
	origfileInfo, err := os.Stat(testFilename)
	require.NoError(t, err)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "unknown line",
		Repl:    "new line",
	}

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.False(t, task.Updated)
	assert.Equal(t, res.Comment, "File not changed")
	assert.NotContains(t, res.Changes, "count")

	currContents := ReadFileContents(t, testFilename)
	currFileInfo, err := os.Stat(testFilename)
	require.NoError(t, err)

	assert.Equal(t, origContents, currContents)
	assert.Equal(t, origfileInfo.ModTime(), currFileInfo.ModTime())
	assert.Equal(t, origfileInfo.Mode(), currFileInfo.Mode())
}

func TestShouldReplaceCountMatchingItems(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "line",
		Repl:    "new line",
		Count:   2,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "File updated")
	assert.Equal(t, res.Changes["count"], "2 replacement(s) made")

	contents := ReadFileContents(t, testFilename)

	matchingCount := strings.Count(contents, "new line")
	assert.Equal(t, 2, matchingCount)
}

func TestShouldSkipWhenFilesizeTooLarge(t *testing.T) {
	ctx := context.Background()

	largerContents := strings.Repeat("01234567890", 1024)

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, largerContents)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:        "replace-1",
		Name:        testFilename,
		Pattern:     "line",
		Repl:        "new line",
		MaxFileSize: "1k",
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.False(t, task.Updated)

	assert.True(t, res.IsSkipped)
	assert.Contains(t, res.SkipReason, "file size is greater than max_file_size")
}

func TestShouldErrorIfTargetNotRegularFile(t *testing.T) {
	ctx := context.Background()

	var testFilename string
	if runtime.GOOS == "windows" {
		testFilename = `c:\windows\temp`
	} else {
		testFilename = "/tmp"
	}

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "line",
		Repl:    "new line",
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.False(t, task.Updated)

	assert.Contains(t, res.Err.Error(), "is not a regular file")
}

func TestShouldAppendNotFoundContent(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:             "replace-1",
		Name:             testFilename,
		Pattern:          "unknown line",
		Repl:             "not gonna match",
		NotFoundContent:  "an extra line",
		AppendIfNotFound: true,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	contents := ReadFileContents(t, testFilename)

	assert.Equal(t, 42, len(contents))

	assert.Equal(t, res.Comment, "File updated")
	assert.Equal(t, res.Changes["count"], "1 addition(s) made")

	index := strings.Index(contents, "an extra line")
	assert.NotEqual(t, -1, index)
	assert.Equal(t, 29, index)
}

func TestShouldPrependNotFoundContent(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:              "replace-1",
		Name:              testFilename,
		Pattern:           "unknown line",
		Repl:              "not gonna match",
		NotFoundContent:   "an extra line",
		PrependIfNotFound: true,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "File updated")
	assert.Equal(t, res.Changes["count"], "1 addition(s) made")

	contents := ReadFileContents(t, testFilename)

	assert.Equal(t, 42, len(contents))

	index := strings.Index(contents, "an extra line")
	assert.NotEqual(t, -1, index)
	assert.Equal(t, 0, index)
}

func TestShouldUseReplContentWhenWhenNoNotFoundContent(t *testing.T) {
	ctx := context.Background()

	testFilename := getTestFilename()

	WriteTestFile(t, testFilename, simpleTestFileContentWithRepetition)
	defer os.Remove(testFilename)

	executor := &FileReplaceTaskExecutor{
		FsManager: &utils.FsManager{},
	}
	task := &FileReplaceTask{
		Path:    "replace-1",
		Name:    testFilename,
		Pattern: "unknown line",
		Repl:    "an extra line",
		// NotFoundContent:  "an extra line",
		AppendIfNotFound: true,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "File updated")
	assert.Equal(t, res.Changes["count"], "1 addition(s) made")

	contents := ReadFileContents(t, testFilename)

	assert.Equal(t, 42, len(contents))

	index := strings.Index(contents, "an extra line")
	assert.NotEqual(t, -1, index)
	assert.Equal(t, 29, index)
}
