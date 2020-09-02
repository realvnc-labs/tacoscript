package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/cloudradar-monitoring/tacoscript/applog"

	"github.com/goftp/server"
	log "github.com/sirupsen/logrus"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	filedriver "github.com/goftp/file-driver"

	"github.com/stretchr/testify/assert"
)

func init() {
	applog.Init(true)
}

type fileManagedTestCase struct {
	FileShouldExist bool
	Name            string
	ContentToWrite  string
	LogExpectation  string
	Task            *FileManagedTask
	ExpectedResult  ExecutionResult
	RunnerMock      *appExec.SystemRunner
	FileExpectation *utils.FileExpectation
	ExpectedCmdStrs []string
}

func TestFileManagedTaskExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ftpPort = 3021

	ftpURL, err := startFTPServer(ctx, ftpPort)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	httpSrvURL, httpSrv, err := startHTTPServer(false)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer httpSrv.Close()

	httpsSrvURL, httpsSrv, err := startHTTPServer(true)
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

	err = ioutil.WriteFile("sourceFileAtLocal.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	err = ioutil.WriteFile("sourceFileHTTPS.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpsSrvURL.Path = "/sourceFileHTTPS.txt"

	err = ioutil.WriteFile("sourceFileHTTP.txt", []byte("one two three"), 0600)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpSrvURL.Path = "/sourceFileHTTP.txt"

	err = ioutil.WriteFile("sourceFileFTP.txt", []byte("one two three"), 0600)
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
				Creates: []string{"some file 123", "some file 345"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
			},
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
			},
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
				Err: errors.New("unknown hash algorithm 'md4' in 'md4=5e4fe0155703dde467f3ab234e6f966f'"),
			},
		},
		{
			Name: "local_source_copy_success",
			Task: &FileManagedTask{
				Path:       "local_source_copy_success_path",
				Name:       "targetFileAtLocal.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
			},
			ExpectedResult: ExecutionResult{},
			FileExpectation: &utils.FileExpectation{
				FilePath:        "targetFileAtLocal.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
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
			FileExpectation: &utils.FileExpectation{
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
				Err: fmt.Errorf(
					"checksum 'md5=dafdfdafdafdfad' didn't match with checksum 'md5=5e4fe0155703dde467f3ab234e6f966f' of the remote source '%s'",
					httpSrvURL.String(),
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
			FileExpectation: &utils.FileExpectation{
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
			FileExpectation: &utils.FileExpectation{
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
						cmdStr := cmd.String()
						if strings.Contains(cmdStr, "cmd with OnlyIf failure") {
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
			FileExpectation: &utils.FileExpectation{
				FilePath:        "onlyIfConditionTrue.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "saving_contents_to_file",
			Task: &FileManagedTask{
				Name: "contentsToFile.txt",
				Path: "saving_contents_to_file",
				Contents: sql.NullString{
					Valid: true,
					String: `one
two
three`,
				},
			},
			ContentToWrite: "one two three",
			ExpectedResult: ExecutionResult{},
			FileExpectation: &utils.FileExpectation{
				FilePath:    "contentsToFile.txt",
				ShouldExist: true,
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
				Name: "contentsToFile.txt",
				Path: "saving_contents_to_file",
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
			FileExpectation: &utils.FileExpectation{
				FilePath:    "contentsToFile2.txt",
				ShouldExist: true,
				ExpectedContent: `one
two
three`,
			},
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
				e := ioutil.WriteFile(tc.Task.Name, []byte(tc.ContentToWrite), 0600)
				assert.NoError(t, e)
			}
			runner := tc.RunnerMock
			if runner == nil {
				runner = &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}}
			}
			fileManagedExecutor := &FileManagedTaskExecutor{
				Runner: runner,
				FsManager: &utils.FsManagerMock{
					ExistsToReturn: tc.FileShouldExist,
				},
			}

			lc.Messages = []string{}

			res := fileManagedExecutor.Execute(context.Background(), tc.Task)

			assertTestCase(tt, &tc, res, lc)
			filesToDelete = append(filesToDelete, tc.Task.Name)
		})
	}

	err = deleteFiles(filesToDelete)
	if err != nil {
		log.Warn(err)
	}
}

func assertTestCase(t *testing.T, tc *fileManagedTestCase, res ExecutionResult, logs *applog.BufferedLogs) {
	assert.EqualValues(t, tc.ExpectedResult.Err, res.Err)
	assert.EqualValues(t, tc.ExpectedResult.IsSkipped, res.IsSkipped)
	assert.EqualValues(t, tc.ExpectedResult.StdOut, res.StdOut)
	assert.EqualValues(t, tc.ExpectedResult.StdErr, res.StdErr)

	var cmds []*exec.Cmd
	if tc.RunnerMock != nil {
		systemAPIMock := tc.RunnerMock.SystemAPI.(*appExec.SystemAPIMock)
		cmds = systemAPIMock.Cmds
	}
	AssertCmdsPartiallyMatch(t, tc.ExpectedCmdStrs, cmds)

	if tc.LogExpectation != "" {
		assertLogExpectation(t, tc.LogExpectation, logs)
	}

	if tc.FileExpectation == nil {
		return
	}

	isExpectationMatched, nonMatchedReason, err := utils.AssertFileMatchesExpectation(tc.Task.Name, tc.FileExpectation)
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

func deleteFiles(files []string) error {
	errs := &utils.Errors{
		Errs: []error{},
	}
	for _, file := range files {
		errs.Add(os.Remove(file))
	}

	return errs.ToError()
}

func startHTTPServer(isHTTPS bool) (u *url.URL, srv *httptest.Server, err error) {
	if isHTTPS {
		srv = httptest.NewTLSServer(http.FileServer(http.Dir(".")))
	} else {
		srv = httptest.NewServer(http.FileServer(http.Dir(".")))
	}

	u, err = url.Parse(srv.URL)

	return
}

func startFTPServer(ctx context.Context, port int) (*url.URL, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ftpHost := fmt.Sprintf("ftp://root:root@localhost:%d", port)
	ftpHostURL, err := url.Parse(ftpHost)
	if err != nil {
		return nil, err
	}

	factory := &filedriver.FileDriverFactory{
		RootPath: path,
		Perm:     server.NewSimplePerm("user", "group"),
	}

	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     port,
		Hostname: "localhost",
		Auth:     &server.SimpleAuth{Name: "root", Password: "root"},
	}

	log.Printf("Starting ftp server on %v:%v", opts.Hostname, opts.Port)
	ftpSrvr := server.NewServer(opts)

	go func() {
		<-ctx.Done()
		err := ftpSrvr.Shutdown()
		if err != nil {
			log.Error(err)
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		err := ftpSrvr.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return ftpHostURL, err
	case <-time.After(time.Millisecond * 300):
		return ftpHostURL, nil
	}
}
