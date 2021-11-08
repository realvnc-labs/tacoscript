package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/applog"

	log "github.com/sirupsen/logrus"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/apptest"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/stretchr/testify/assert"
)

func init() {
	applog.Init(true)
}

type fileManagedTestCase struct {
	Name                   string
	ContentToWrite         string
	ContentEncodingToWrite string
	LogExpectation         string
	Task                   *FileManagedTask
	ExpectedResult         ExecutionResult
	RunnerMock             *appExec.SystemRunner
	FileExpectation        *apptest.FileExpectation
	ExpectedCmdStrs        []string
	ErrorExpectation       *apptest.ErrorExpectation
}

func TestFileManagedTaskExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ftpPort = 3021

	ftpURL, err := apptest.StartFTPServer(ctx, ftpPort, time.Millisecond*300)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	httpSrvURL, httpSrv, err := apptest.StartHTTPServer(false)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer httpSrv.Close()

	httpsSrvURL, httpsSrv, err := apptest.StartHTTPServer(true)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer httpsSrv.Close()

	filesToDelete := []string{
		"sourceFileHTTPS.txt",
		"sourceFileHTTP.txt",
		"sourceFileAtLocal.txt",
		"sourceFileFTP.txt",
	}

	err = os.WriteFile("sourceFileAtLocal.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	err = os.WriteFile("sourceFileHTTPS.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpsSrvURL.Path = "/sourceFileHTTPS.txt"

	err = os.WriteFile("sourceFileHTTP.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpSrvURL.Path = "/sourceFileHTTP.txt"

	err = os.WriteFile("sourceFileFTP.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	ftpURL.Path = "sourceFileFTP.txt"

	testCases := []fileManagedTestCase{
		{
			Name: "test creates field",
			Task: &FileManagedTask{
				Path:    "somepath",
				Name:    "some test command",
				Creates: []string{"some_file.123", "sourceFileAtLocal.txt"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
		},
		{
			Name: "test hash matches",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
			ContentToWrite: "one two three",
		},
		{
			Name: "test_wrong_hash_format_error",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md4=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				Err: errors.New("unknown hash algorithm 'md4'"),
			},
		},
		{
			Name: "local_source_copy_success",
			Task: &FileManagedTask{
				Path:       "local_source_copy_success_path",
				Name:       "targetFileAtLocalCopy.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
				Mode: 0777,
			},
			ExpectedResult: ExecutionResult{},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "targetFileAtLocalCopy.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
				ExpectedMode:    0777,
			},
		},
		{
			Name: "http_source_copy_success",
			Task: &FileManagedTask{
				Path:       "http_source_copy_success_path",
				Name:       "targetFileFromHttp.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					URL:         httpSrvURL,
					RawLocation: httpSrvURL.String(),
				},
			},
			ExpectedResult: ExecutionResult{},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "targetFileFromHttp.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "http_source_copy_wrong_source_checksum",
			Task: &FileManagedTask{
				Path:       "http_source_copy_wrong_source_checksum_path",
				Name:       "targetFileFromHttp2.txt",
				SourceHash: "md5=dafdfdafdafdfad",
				Source: utils.Location{
					IsURL:       true,
					URL:         httpSrvURL,
					RawLocation: httpSrvURL.String(),
				},
			},
			ExpectedResult: ExecutionResult{
				Err: errors.New(
					"expected hash sum 'md5=dafdfdafdafdfad' didn't match with " +
						"checksum 'md5=5e4fe0155703dde467f3ab234e6f966f' of the source file " +
						"'targetFileFromHttp2.txt_temp'",
				),
			},
		},
		{
			Name: "https_source_copy_success",
			Task: &FileManagedTask{
				Path:       "https_source_copy_success_path",
				Name:       "targetFileFromHttps.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					URL:         httpsSrvURL,
					RawLocation: httpsSrvURL.String(),
				},
				SkipTLSCheck: true,
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "targetFileFromHttps.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "ftp_source_copy_success",
			Task: &FileManagedTask{
				Path:       "ftp_source_copy_success_path",
				Name:       "targetFileFromFtp.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					URL:         ftpURL,
					RawLocation: ftpURL.String(),
				},
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "targetFileFromFtp.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "executing_onlyif_condition_failure",
			Task: &FileManagedTask{
				Name:   "cmd with OnlyIf failure",
				OnlyIf: []string{"check OnlyIf error"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
			ExpectedCmdStrs: []string{"check OnlyIf error"},
			RunnerMock: &appExec.SystemRunner{
				SystemAPI: &appExec.SystemAPIMock{
					Cmds: []*exec.Cmd{},
					Callback: func(cmd *exec.Cmd) error {
						if strings.Contains(cmd.String(), "cmd with OnlyIf failure") {
							return nil
						}
						return appExec.RunError{Err: errors.New("some OnlyIfFailure")}
					},
				},
			},
		},
		{
			Name: "executing_onlyif_condition_success",
			Task: &FileManagedTask{
				Name:   "onlyIfConditionTrue.txt",
				OnlyIf: []string{"check OnlyIf success 1", "check OnlyIf success 2"},
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult:  ExecutionResult{},
			ExpectedCmdStrs: []string{"check OnlyIf success 1", "check OnlyIf success 2"},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "onlyIfConditionTrue.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "saving_contents_to_file",
			Task: &FileManagedTask{
				Name: "contentsToFile.txt",
				Path: "saving_contents_to_file_path",
				Contents: sql.NullString{
					Valid: true,
					String: `one
two
three`,
				},
				Replace: true,
				Mode:    0777,
			},
			ContentToWrite: "one two three",
			ExpectedResult: ExecutionResult{},
			FileExpectation: &apptest.FileExpectation{
				FilePath:     "contentsToFile.txt",
				ShouldExist:  true,
				ExpectedMode: 0777,
				ExpectedContent: `one
two
three`,
			},
			LogExpectation: `-one
-two
-three
+one two three
`,
		},
		{
			Name: "skipping_content_on_empty_diff",
			Task: &FileManagedTask{
				Name: "skipping_content_on_empty_diff.txt",
				Path: "skipping_content_on_empty_diff_file",
				Contents: sql.NullString{
					Valid: true,
					String: `one
two
three`,
				},
			},
			ContentToWrite: `one
two
three`,
			ExpectedResult: ExecutionResult{IsSkipped: true},
			FileExpectation: &apptest.FileExpectation{
				FilePath:    "skipping_content_on_empty_diff.txt",
				ShouldExist: true,
				ExpectedContent: `one
two
three`,
			},
		},
		{
			Name: "make_dirs_success",
			Task: &FileManagedTask{
				MakeDirs:   true,
				Name:       "sub/dir/sourceFileAtLocal.txt",
				Path:       "make_dirs_success_path",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
			},
			ExpectedResult: ExecutionResult{},
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "sub/dir/sourceFileAtLocal.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
			},
		},
		{
			Name: "make_dirs_fail",
			Task: &FileManagedTask{
				MakeDirs:   false,
				Name:       "dfasdfaf/sourceFileAtLocal2.txt",
				Path:       "make_dirs_fail_path",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
			},
			FileExpectation: &apptest.FileExpectation{
				FilePath:    "dfasdfaf/sourceFileAtLocal2.txt",
				ShouldExist: false,
			},
			ErrorExpectation: &apptest.ErrorExpectation{
				PartialText: "dfasdfaf/sourceFileAtLocal2.txt",
			},
		},
		{
			Name: "replace_false",
			Task: &FileManagedTask{
				Replace:    false,
				Name:       "existingFileToReplace.txt",
				Path:       "replace_false_path",
				SourceHash: "md5=111",
				Contents: sql.NullString{
					String: "content to ignore",
					Valid:  true,
				},
				Mode: 0777,
			},
			ContentToWrite: "one two three",
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "existingFileToReplace.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
				ExpectedMode:    0777,
			},
		},
		{
			Name: "replace_true",
			Task: &FileManagedTask{
				Replace:    true,
				Name:       "existingFileToReplace2.txt",
				Path:       "replace_true_path",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Contents: sql.NullString{
					String: "one two three",
					Valid:  true,
				},
			},
			ContentToWrite: "one",
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "existingFileToReplace2.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
			},
		},
		{
			Name: "skip_verify_success_for_url",
			Task: &FileManagedTask{
				SkipVerify: true,
				Replace:    true,
				Name:       "skipVerifyFileSuccess.txt",
				Path:       "skip_verify_success_for_url_path",
				Source: utils.Location{
					IsURL:       true,
					URL:         httpSrvURL,
					RawLocation: httpSrvURL.String(),
				},
			},
			ContentToWrite: " ",
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "skipVerifyFileSuccess.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
			},
		},
		{
			Name: "skip_verify_no_content_change",
			Task: &FileManagedTask{
				SkipVerify: true,
				Replace:    true,
				Name:       "skipVerifyFileNoChange.txt",
				Path:       "skip_verify_no_content_change_path",
				Source: utils.Location{
					IsURL:       true,
					URL:         ftpURL,
					RawLocation: ftpURL.String(),
				},
			},
			ContentToWrite: "one two three",
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "skipVerifyFileNoChange.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
			},
		},
		{
			Name: "skip_verify_success_for_local",
			Task: &FileManagedTask{
				SkipVerify: true,
				Replace:    true,
				Name:       "skipVerifyFileLocalSuccess.txt",
				Path:       "skip_verify_success_for_local_path",
				Source: utils.Location{
					IsURL:       true,
					URL:         httpSrvURL,
					RawLocation: httpSrvURL.String(),
				},
				Mode: 0777,
			},
			ContentToWrite: " ",
			FileExpectation: &apptest.FileExpectation{
				FilePath:        "skipVerifyFileLocalSuccess.txt",
				ShouldExist:     true,
				ExpectedContent: `one two three`,
				ExpectedMode:    0777,
			},
		},
		{
			Name: "encoding_success",
			Task: &FileManagedTask{
				Replace:    true,
				SkipVerify: true,
				Name:       "encodingSuccessFile.txt",
				Path:       "encoding_success_path",
				Contents: sql.NullString{
					String: "一些中文内容",
					Valid:  true,
				},
				Encoding: "gb18030",
			},
			FileExpectation: &apptest.FileExpectation{
				FilePath:         "encodingSuccessFile.txt",
				ShouldExist:      true,
				ExpectedContent:  `一些中文内容`,
				ExpectedEncoding: "gb18030",
			},
		},
		{
			Name: "encoding_with_content_compare",
			Task: &FileManagedTask{
				Name: "encodingWithContentCompare.txt",
				Path: "encoding_with_content_compare_path",
				Contents: sql.NullString{
					Valid:  true,
					String: `一些中文内容`,
				},
				Replace:  true,
				Mode:     0777,
				Encoding: "gb18030",
			},
			ContentToWrite:         "一些中文内",
			ContentEncodingToWrite: "gb18030",
			ExpectedResult:         ExecutionResult{},
			FileExpectation: &apptest.FileExpectation{
				FilePath:         "encodingWithContentCompare.txt",
				ShouldExist:      true,
				ExpectedMode:     0777,
				ExpectedContent:  `一些中文内容`,
				ExpectedEncoding: "gb18030",
			},
			LogExpectation: `-一些中文内容
+一些中文内
`,
		},
	}

	logsCollection := &applog.BufferedLogs{
		Messages: []string{},
	}
	log.AddHook(logsCollection)

	for _, testCase := range testCases {
		tc := testCase
		lc := logsCollection
		t.Run(tc.Name, func(tt *testing.T) {
			if tc.ContentToWrite != "" {
				var e error
				if tc.ContentEncodingToWrite != "" {
					e = utils.WriteEncodedFile(tc.ContentEncodingToWrite, tc.ContentToWrite, tc.Task.Name, 0600)
				} else {
					e = os.WriteFile(tc.Task.Name, []byte(tc.ContentToWrite), 0600)
				}
				assert.NoError(t, e)
			}
			runner := tc.RunnerMock
			if runner == nil {
				runner = &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}}
			}

			fileManagedExecutor := &FileManagedTaskExecutor{
				Runner:      runner,
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			}

			lc.Messages = []string{}

			filesToDelete = append(filesToDelete, tc.Task.Name)

			res := fileManagedExecutor.Execute(context.Background(), tc.Task)

			assertTestCase(tt, &tc, &res, lc)
		})
	}

	err = apptest.DeleteFiles(filesToDelete)
	if err != nil {
		log.Warn(err)
	}
}

func assertTestCase(t *testing.T, tc *fileManagedTestCase, res *ExecutionResult, logs *applog.BufferedLogs) {
	if tc.ErrorExpectation != nil {
		apptest.AssertErrorExpectation(t, res.Err, tc.ErrorExpectation)
	} else {
		if tc.ExpectedResult.Err == nil {
			assert.NoError(t, res.Err)
		} else {
			assert.EqualError(t, res.Err, tc.ExpectedResult.Err.Error())
		}
	}

	assert.EqualValues(t, tc.ExpectedResult.IsSkipped, res.IsSkipped)
	assert.EqualValues(t, tc.ExpectedResult.StdOut, res.StdOut)
	assert.EqualValues(t, tc.ExpectedResult.StdErr, res.StdErr)

	if tc.LogExpectation != "" {
		assertLogExpectation(t, tc.LogExpectation, logs)
	}

	if tc.FileExpectation == nil {
		return
	}

	isExpectationMatched, nonMatchedReason, err := apptest.AssertFileMatchesExpectation(
		tc.FileExpectation,
	)

	assert.NoError(t, err)
	if err != nil {
		return
	}

	if !isExpectationMatched {
		assert.Fail(t, nonMatchedReason)
	}
}

func assertLogExpectation(t *testing.T, expectedLog string, logs *applog.BufferedLogs) {
	for _, msg := range logs.Messages {
		if strings.Contains(msg, expectedLog) {
			return
		}
	}

	assert.Failf(
		t,
		"failed log expectation",
		"didn't find expected log '%s' in the actual logs '%s'",
		expectedLog,
		logs.Messages,
	)
}

func TestFileManagedUserAndGroup(t *testing.T) {
	testCases := []struct {
		task               *FileManagedTask
		chownError         string
		fileStatError      string
		expectedError      string
		expectedChownInput *apptest.ChownInput
	}{
		{
			task: &FileManagedTask{
				Name:       "someFile.txt",
				SkipVerify: true,
				Replace:    true,
				Contents: sql.NullString{
					Valid:  true,
					String: `onetwothree`,
				},
				User:  "someuser",
				Group: "somegroup",
			},
			expectedChownInput: &apptest.ChownInput{
				TargetFilePath: "someFile.txt",
				UserName:       "someuser",
				GroupName:      "somegroup",
			},
		},
		{
			task: &FileManagedTask{
				Name:       "someFile2.txt",
				SkipVerify: true,
				Replace:    true,
				Contents: sql.NullString{
					Valid:  true,
					String: `onetwothree`,
				},
				User:  "someuser",
				Group: "somegroup",
			},
			expectedChownInput: &apptest.ChownInput{
				TargetFilePath: "someFile2.txt",
				UserName:       "someuser",
				GroupName:      "somegroup",
			},
			chownError:    "some chown error",
			expectedError: "some chown error",
		},
		{
			task: &FileManagedTask{
				Name:       "someFile3.txt",
				SkipVerify: true,
				Replace:    true,
				Contents: sql.NullString{
					Valid:  true,
					String: `onetwothree`,
				},
				User:  "someuser",
				Group: "",
			},
			fileStatError: "some stat error",
			expectedError: "some stat error",
		},
	}

	for _, testCase := range testCases {
		fsMock := &apptest.FsManagerMock{
			ChownInputs: []apptest.ChownInput{},
			StatOutputFileInfo: &apptest.FakeFile{
				FileMode: 0644,
			},
		}
		if testCase.chownError != "" {
			fsMock.ChownErrorToReturn = errors.New(testCase.chownError)
		}

		if testCase.fileStatError != "" {
			fsMock.StatOutputError = errors.New(testCase.fileStatError)
		}

		fileManagedExecutor := &FileManagedTaskExecutor{
			FsManager:   fsMock,
			HashManager: &utils.HashManager{},
		}

		res := fileManagedExecutor.Execute(context.Background(), testCase.task)
		assert.False(t, res.IsSkipped)

		if testCase.expectedError != "" {
			assert.EqualError(t, res.Err, testCase.expectedError)
		} else {
			assert.NoError(t, res.Err)
		}

		if testCase.expectedChownInput != nil {
			assert.Len(t, fsMock.ChownInputs, 1)
			if len(fsMock.ChownInputs) == 0 {
				return
			}

			actualChownInput := fsMock.ChownInputs[0]
			assert.EqualValues(t, *testCase.expectedChownInput, actualChownInput)
		}
	}
}

func TestFileManagedTaskValidation(t *testing.T) {
	testCases := []struct {
		Name          string
		ExpectedError string
		Task          FileManagedTask
	}{
		{
			Name: "missing_name",
			Task: FileManagedTask{
				Path:   "somepath",
				Source: utils.Location{RawLocation: "some location"},
			},
			ExpectedError: fmt.Sprintf("empty required value at path 'somepath.%s'", NameField),
		},
		{
			Name: "missing_hash_for_url",
			Task: FileManagedTask{
				Name: "task1",
				Path: "somepath1",
				Source: utils.Location{
					IsURL: true,
					URL: &url.URL{
						Scheme: "http",
						Host:   "ya.ru",
					},
					RawLocation: "http://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(
				`empty '%s' field at path 'somepath1.%s' for remote url source 'http://ya.ru'`,
				SourceHashField,
				SourceHashField,
			),
		},
		{
			Name: "missing_hash_for_ftp",
			Task: FileManagedTask{
				Name: "task2",
				Path: "somepath2",
				Source: utils.Location{
					IsURL: true,
					URL: &url.URL{
						Scheme: "ftp",
						Host:   "ya.ru",
					},
					RawLocation: "ftp://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(
				`empty '%s' field at path 'somepath2.%s' for remote url source 'ftp://ya.ru'`,
				SourceHashField,
				SourceHashField,
			),
		},
		{
			Name: "valid_task",
			Task: FileManagedTask{
				Name: "some p",
				Source: utils.Location{
					RawLocation: "some location",
				},
			},
		},
		{
			Name: "missing_source_and_content",
			Task: FileManagedTask{
				Name: "task missing source",
				Path: "some task missing source",
			},
			ExpectedError: `either content or source should be provided for the task at path 'some task missing source'`,
		},
		{
			Name: "empty_content",
			Task: FileManagedTask{
				Name: "task empty_content",
				Contents: sql.NullString{
					Valid:  true,
					String: "",
				},
			},
		},
		{
			Name: "missing_hash_with_skip_verify",
			Task: FileManagedTask{
				Name: "missing_hash_with_skip_verify",
				Path: "missing_hash_with_skip_verify_path",
				Source: utils.Location{
					IsURL:       true,
					RawLocation: "ftp://ya.ru",
				},
				SkipVerify: true,
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Task.Validate()
			if tc.ExpectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.ExpectedError)
			}
		})
	}
}
