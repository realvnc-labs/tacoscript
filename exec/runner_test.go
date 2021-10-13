package exec

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/stretchr/testify/assert"
)

func TestOSApi(t *testing.T) {
	cmd := exec.Command("echo", "123")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	cmdRunner := OSApi{}
	err := cmdRunner.Run(cmd)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, "123\n", outBuf.String())
}

func TestRunner(t *testing.T) {
	testCases := []struct {
		execContext    *Context
		name           string
		setUserError   string
		cmdExecError   string
		expectedErr    error
		expectedStdOut string
		expectedStdErr string
	}{
		{
			name: "run with all fields filled",
			execContext: &Context{
				StdoutWriter: &bytes.Buffer{},
				StderrWriter: &bytes.Buffer{},
				WorkingDir:   "some workdir",
				User:         "some user",
				Path:         "some path",
				Envs: conv.KeyValues{
					{
						Key:   "1",
						Value: "one",
					},
					{
						Key:   "2",
						Value: "two",
					},
				},
				Cmds:  []string{"cmd1", "cmd2"},
				Shell: "shell",
			},
			expectedStdOut: "some stdout",
			expectedStdErr: "some stderr",
		},
		{
			name: "test adding c shell param",
			execContext: &Context{
				StdoutWriter: &bytes.Buffer{},
				StderrWriter: &bytes.Buffer{},
				Cmds:         []string{"cmd3", "cmd4"},
				Shell:        "zsh",
			},
		},
		{
			name: "default shell",
			execContext: &Context{
				StdoutWriter: &bytes.Buffer{},
				StderrWriter: &bytes.Buffer{},
				Cmds:         []string{"somecmd"},
				Shell:        "",
			},
		},
		{
			name:         "test user set failure",
			setUserError: "set user failed",
			expectedErr:  errors.New("set user failed"),
			execContext: &Context{
				Cmds: []string{"cmd5"},
				User: "some failing user",
			},
		},
		{
			name:         "test base cmd error",
			cmdExecError: "cmd exec failed",
			expectedErr:  RunError{Err: errors.New("cmd exec failed")},
			execContext: &Context{
				StdoutWriter: &bytes.Buffer{},
				StderrWriter: &bytes.Buffer{},
				Cmds:         []string{"cmd6"},
			},
		},
		{
			name: "test trimming space of cmds",
			execContext: &Context{
				StdoutWriter: &bytes.Buffer{},
				StderrWriter: &bytes.Buffer{},
				Cmds:         []string{"cmd7   ", "     cmd8"},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			systemAPI := &SystemAPIMock{
				Cmds: []*exec.Cmd{},
			}

			if tc.setUserError != "" {
				systemAPI.UserSetErrToReturn = errors.New(tc.setUserError)
			}

			if tc.cmdExecError != "" {
				systemAPI.ErrToGive = errors.New(tc.cmdExecError)
			}

			systemRunner := SystemRunner{
				SystemAPI: systemAPI,
			}

			execContext := tc.execContext
			err := systemRunner.Run(execContext)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			assert.NoError(t, err)
			if err != nil {
				return
			}

			if execContext.User != "" {
				assert.Equal(t, execContext.User, systemAPI.UserNameInput)
			}

			if execContext.Path != "" {
				assert.Equal(t, execContext.Path, systemAPI.UserNamePathInput)
			}

			actuallyExecutedCmds := systemAPI.Cmds
			for _, actuallyExecutedCmd := range actuallyExecutedCmds {
				assert.Equal(t, execContext.WorkingDir, actuallyExecutedCmd.Dir)

				if len(execContext.Envs) > 0 {
					envsGiven := execContext.Envs.ToEqualSignStrings()
					for _, envGiven := range envsGiven {
						givenEnvFound := false
						for _, actualEnv := range actuallyExecutedCmd.Env {
							if actualEnv == envGiven {
								givenEnvFound = true
								break
							}
						}
						if !givenEnvFound {
							assert.Failf(
								t,
								"cannot find an env variable",
								"expected variable '%s', actual environment variables: %+v",
								envGiven,
								actuallyExecutedCmd.Env,
							)
						}
					}
				}

				if tc.expectedStdOut != "" {
					_, err := actuallyExecutedCmd.Stdout.Write([]byte(tc.expectedStdOut))
					assert.NoError(t, err)
					actualStdOut := execContext.StdoutWriter.(*bytes.Buffer).String()
					assert.Contains(t, actualStdOut, tc.expectedStdOut)
				}

				if tc.expectedStdErr != "" {
					_, err := actuallyExecutedCmd.Stderr.Write([]byte(tc.expectedStdErr))
					assert.NoError(t, err)
					actualStdErr := execContext.StderrWriter.(*bytes.Buffer).String()
					assert.Contains(t, actualStdErr, tc.expectedStdErr)
				}
			}
		})
	}
}
