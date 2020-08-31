package tasks

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/stretchr/testify/assert"
)

type fileManagedTestCase struct {
	Task            *FileManagedTask
	ExpectedResult  ExecutionResult
	RunnerMock      *appExec.SystemRunner
	Name            string
	FileShouldExist bool
	ContentToWrite  string
	FileExpectation *utils.FileExpectation
}

func TestFileManagedTaskExecution(t *testing.T) {
	filesToDelete := make([]string, 0)

	testCases := []struct {
		Task            *FileManagedTask
		ExpectedResult  ExecutionResult
		RunnerMock      *appExec.SystemRunner
		Name            string
		FileShouldExist bool
		ContentToWrite  string
		FileExpectation *utils.FileExpectation
	}{
		{
			Name: "test creates field",
			Task: &FileManagedTask{
				Path:    "somepath",
				Name:    "some test command",
				Creates: []string{"some file 123", "some file 345"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock:      &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}},
			FileShouldExist: true,
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
				Err:       nil,
			},
			RunnerMock:     &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}},
			ContentToWrite: "one two three",
		},
		{
			Name: "test wrong hash format error",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md4=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("unknown hash algorithm 'md4' in 'md4=5e4fe0155703dde467f3ab234e6f966f'"),
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}},
		},
		{
			Name: "local_source_copy_success",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md4=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("unknown hash algorithm 'md4' in 'md4=5e4fe0155703dde467f3ab234e6f966f'"),
			},
			RunnerMock: &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			assertTestCase(tt, tc)
			filesToDelete = append(filesToDelete, tc.Task.Name)
		})
	}

	err := deleteFiles(filesToDelete)
	if err != nil {
		log.Warn(err)
	}
}

func assertTestCase(t *testing.T, tc fileManagedTestCase) {
	if tc.ContentToWrite != "" {
		err := ioutil.WriteFile(tc.Task.Name, []byte(tc.ContentToWrite), 0644)
		assert.NoError(t, err)
	}

	fileManagedExecutor := &FileManagedTaskExecutor{
		Runner: tc.RunnerMock,
		FsManager: &utils.FsManagerMock{
			ExistsToReturn: tc.FileShouldExist,
		},
	}

	res := fileManagedExecutor.Execute(context.Background(), tc.Task)
	assert.EqualValues(t, tc.ExpectedResult.Err, res.Err)
	assert.EqualValues(t, tc.ExpectedResult.IsSkipped, res.IsSkipped)
	assert.EqualValues(t, tc.ExpectedResult.StdOut, res.StdOut)
	assert.EqualValues(t, tc.ExpectedResult.StdErr, res.StdErr)

	if tc.ExpectedResult.Err != nil {
		return
	}

	if tc.ExpectedResult.IsSkipped {
		return
	}

	if tc.FileExpectation == nil {
		return
	}

	isExpectationMatched, nonMatchedReason, err := utils.AssertFileMatchesExpectation(tc.Name, tc.FileExpectation)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	if !isExpectationMatched {
		assert.Fail(t, nonMatchedReason)
	}
}

func TestFileManagedTaskValidation(t *testing.T) {
	testCases := []struct {
		Task          FileManagedTask
		ExpectedError string
	}{
		{
			Task: FileManagedTask{
				Path: "somepath",
			},
			ExpectedError: fmt.Sprintf("empty required value at path 'somepath.%s'", NameField),
		},
		{
			Task: FileManagedTask{
				Name: "task1",
				Path: "somepath1",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
						Scheme:  "http",
						Host:    "ya.ru",
					},
					RawLocation: "http://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(`empty '%s' field at path 'somepath1.%s' for remote url source 'http://ya.ru'`, SourceHashField, SourceHashField),
		},
		{
			Task: FileManagedTask{
				Name: "task2",
				Path: "somepath2",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
						Scheme: "ftp",
						Host:   "ya.ru",
					},
					RawLocation:  "ftp://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(`empty '%s' field at path 'somepath2.%s' for remote url source 'ftp://ya.ru'`, SourceHashField, SourceHashField),
		},
		{
			Task: FileManagedTask{
				Name: "some p",
				Source: utils.Location{
					IsURL:     false,
					LocalPath: "/somepath",
				},
			},
		},
	}

	for _, testCase := range testCases {
		err := testCase.Task.Validate()
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
	}
}

func deleteFiles(files []string) error {
	errs := &utils.Errors{
		Errs: []error{},
	}
	for _, file := range files {
		errs.Add(os.Remove(file))
	}

	return errs.ToError()
}
