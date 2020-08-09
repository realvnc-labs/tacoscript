package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/conv"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RunnerMock struct {
	cmds      []*exec.Cmd
	errToGive error
	id        string
}

func (r *RunnerMock) Run(cmd *exec.Cmd) error {
	r.cmds = append(r.cmds, cmd)

	return r.errToGive
}

type UserSystemInfoParserMock struct {
	userNameInput      string
	pathInput          string
	sysUserIDToReturn  uint32
	sysGroupIDToReturn uint32
	errToReturn        error
}

func (usipm *UserSystemInfoParserMock) Parse(userName, path string) (sysUserID, sysGroupID uint32, err error) {
	usipm.userNameInput = userName
	usipm.pathInput = path

	return usipm.sysUserIDToReturn, usipm.sysGroupIDToReturn, usipm.errToReturn
}

func TestOSCmdRunner(t *testing.T) {
	cmd := exec.Command("echo", "123")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	cmdRunner := OSCmdRunner{}
	err := cmdRunner.Run(cmd)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, "123\n", outBuf.String())
}

func TestCmdRunTaskBuilder(t *testing.T) {
	runnerMock := &RunnerMock{
		cmds: []*exec.Cmd{},
		id:   "some id",
	}

	testCases := []struct {
		typeName     string
		path         string
		ctx          []map[string]interface{}
		expectedTask *CmdRunTask
	}{
		{
			typeName: "someType",
			path:     "somePath",
			ctx: []map[string]interface{}{
				{
					NameField:  1,
					CwdField:   "somedir",
					UserField:  "someuser",
					ShellField: "someshell",
					EnvField: buildExpectedEnvs(map[string]interface{}{
						"one": "1",
						"two": "2",
					}),
					CreatesField: "somefile.txt",
				},
			},
			expectedTask: &CmdRunTask{
				TypeName:   "someType",
				Path:       "somePath",
				Cmd:        "1",
				WorkingDir: "somedir",
				User:       "someuser",
				Shell:      "someshell",
				Envs: conv.KeyValues{
					{
						Key:   "one",
						Value: "1",
					},
					{
						Key:   "two",
						Value: "2",
					},
				},
				MissingFileCondition: "somefile.txt",
				Runner:               runnerMock,
				Errors:               &ValidationErrors{},
			},
		},
		{
			typeName: "someTypeWithErrors",
			path:     "somePathWithErrors",
			ctx: []map[string]interface{}{
				{
					EnvField: 123,
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors",
				Path:     "somePathWithErrors",
				Envs:     conv.KeyValues{},
				Runner:   runnerMock,
				Errors: &ValidationErrors{
					Errs: []error{
						fmt.Errorf("wrong env variables value: array is exected at path somePathWithErrors.env but got 123"),
					},
				},
			},
		},
		{
			typeName: "someTypeWithErrors2",
			path:     "somePathWithErrors2",
			ctx: []map[string]interface{}{
				{
					EnvField: []interface{}{
						"one",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors2",
				Path:     "somePathWithErrors2",
				Envs:     conv.KeyValues{},
				Runner:   runnerMock,
				Errors: &ValidationErrors{
					Errs: []error{
						fmt.Errorf("wrong env variables value: array of scalar values is exected at path somePathWithErrors2.env but got [one]"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		cmdBuilder := CmdRunTaskBuilder{
			Runner: runnerMock,
		}
		actualTask, err := cmdBuilder.Build(
			testCase.typeName,
			testCase.path,
			testCase.ctx,
		)

		assert.NoError(t, err)
		if err != nil {
			continue
		}

		actualCmdRunTask, ok := actualTask.(*CmdRunTask)
		assert.True(t, ok)
		if !ok {
			continue
		}

		assert.Equal(t, testCase.expectedTask.User, actualCmdRunTask.User)
		AssertEnvValuesMatch(t, testCase.expectedTask.Envs, actualCmdRunTask.Envs.ToEqualSignStrings())
		assert.Equal(t, testCase.expectedTask.Path, actualCmdRunTask.Path)
		assert.Equal(t, testCase.expectedTask.WorkingDir, actualCmdRunTask.WorkingDir)
		assert.Equal(t, testCase.expectedTask.MissingFileCondition, actualCmdRunTask.MissingFileCondition)
		assert.Equal(t, testCase.expectedTask.Cmd, actualCmdRunTask.Cmd)
		assert.Equal(t, testCase.expectedTask.TypeName, actualCmdRunTask.TypeName)
		assert.Equal(t, testCase.expectedTask.Shell, actualCmdRunTask.Shell)
	}
}

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		Task                            *CmdRunTask
		ExpectedResult                  ExecutionResult
		RunnerMock                      *RunnerMock
		UserSystemInfoParserMock        *UserSystemInfoParserMock
		ExpectedCmdStr                  string
		ShouldCreateFileForMissingCheck bool
	}{
		{
			Task: &CmdRunTask{
				Path:       "somepath",
				Cmd:        "ls -la",
				WorkingDir: "/tmp/dev",
				User:       "user",
				Shell:      "zsh",
				Envs: conv.KeyValues{
					{
						Key:   "someenvkey1",
						Value: "someenvval2",
					},
				},
				MissingFileCondition: "",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       nil,
			},
			ExpectedCmdStr: "zsh -c ls -la",
			RunnerMock: &RunnerMock{
				cmds:      []*exec.Cmd{},
				errToGive: nil,
				id:        "some id",
			},
			UserSystemInfoParserMock: &UserSystemInfoParserMock{
				sysUserIDToReturn:  10,
				sysGroupIDToReturn: 12,
				errToReturn:        nil,
			},
		},
		{
			UserSystemInfoParserMock: &UserSystemInfoParserMock{
				errToReturn: errors.New("some error"),
			},
			Task: &CmdRunTask{
				User:                 "some user",
				Cmd:                  "ls -la",
				MissingFileCondition: "somefile.txt",
			},
			ShouldCreateFileForMissingCheck: true,
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			RunnerMock: &RunnerMock{
				cmds:      []*exec.Cmd{},
				errToGive: nil,
				id:        "some id",
			},
		},
		{
			UserSystemInfoParserMock: &UserSystemInfoParserMock{
				errToReturn: errors.New("some error"),
			},
			Task: &CmdRunTask{
				Cmd:  "echo 12345",
				User: "some user",
			},
			ShouldCreateFileForMissingCheck: false,
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some error"),
			},
			RunnerMock: &RunnerMock{
				cmds:      []*exec.Cmd{},
				errToGive: nil,
				id:        "some id",
			},
		},
		{
			UserSystemInfoParserMock: &UserSystemInfoParserMock{
				errToReturn: nil,
			},
			Task: &CmdRunTask{
				Cmd: "lpwd",
			},
			ShouldCreateFileForMissingCheck: false,
			ExpectedResult: ExecutionResult{
				IsSkipped: false,
				Err:       errors.New("some runner error"),
			},
			RunnerMock: &RunnerMock{
				cmds:      []*exec.Cmd{},
				errToGive: errors.New("some runner error"),
				id:        "some id",
			},
		},
	}

	for _, testCase := range testCases {
		testCase.Task.Runner = testCase.RunnerMock
		testCase.Task.UserSystemInfoParser = testCase.UserSystemInfoParserMock

		if testCase.ShouldCreateFileForMissingCheck && testCase.Task.MissingFileCondition != "" {
			emptyFile, err := os.Create(testCase.Task.MissingFileCondition)
			assert.NoError(t, err)
			if err != nil {
				continue
			}

			err = emptyFile.Close()
			assert.NoError(t, err)
		}

		res := testCase.Task.Execute(context.Background())
		assert.EqualValues(t, testCase.ExpectedResult.Err, res.Err)
		assert.EqualValues(t, testCase.ExpectedResult.IsSkipped, res.IsSkipped)

		cmds := testCase.RunnerMock.cmds

		if testCase.ExpectedResult.IsSkipped {
			assert.Len(t, cmds, 0)
			continue
		}

		if testCase.ExpectedResult.Err != nil {
			continue
		}

		assert.Len(t, cmds, 1)
		cmd := cmds[0]
		assert.Contains(t, cmd.String(), testCase.ExpectedCmdStr)
		assert.Equal(t, testCase.Task.WorkingDir, cmd.Dir)

		if testCase.Task.User != "" {
			assert.Equal(t, true, cmd.SysProcAttr.Setpgid)
			assert.Equal(t, testCase.UserSystemInfoParserMock.sysUserIDToReturn, cmd.SysProcAttr.Credential.Uid)
			assert.Equal(t, testCase.UserSystemInfoParserMock.sysGroupIDToReturn, cmd.SysProcAttr.Credential.Gid)
		}

		AssertEnvValuesMatch(t, testCase.Task.Envs, cmd.Env)

		assert.Equal(t, testCase.Task.User, testCase.UserSystemInfoParserMock.userNameInput)
		assert.Equal(t, testCase.Task.Path, testCase.UserSystemInfoParserMock.pathInput)

		if testCase.ShouldCreateFileForMissingCheck && testCase.Task.MissingFileCondition != "" {
			err := os.Remove(testCase.Task.MissingFileCondition)
			assert.NoError(t, err)
		}
	}
}

func AssertEnvValuesMatch(t *testing.T, expectedEnvs conv.KeyValues, actualCmdEnvs []string) {
	expectedRawEnvs := expectedEnvs.ToEqualSignStrings()
	notFoundEnvs := make([]string, 0, len(expectedEnvs))
	for _, expectedRawEnv := range expectedRawEnvs {
		foundEnv := false
		for _, actualCmdEnv := range actualCmdEnvs {
			if expectedRawEnv == actualCmdEnv {
				foundEnv = true
				break
			}
		}

		if !foundEnv {
			notFoundEnvs = append(notFoundEnvs, expectedRawEnv)
		}
	}

	assert.Empty(
		t,
		notFoundEnvs,
		"was not able to find expected environment variables %s in cmd envs %s",
		strings.Join(notFoundEnvs, ", "),
		strings.Join(actualCmdEnvs, ", "),
	)
}
