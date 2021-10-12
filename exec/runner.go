package exec

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	io2 "github.com/cloudradar-monitoring/tacoscript/io"
	"github.com/sirupsen/logrus"
)

const defaultWindowsShell = "cmd.exe"
const defaultUnixShell = "sh"

type SystemAPI interface {
	Run(cmd *exec.Cmd) error
	SetUser(userName, path string, cmd *exec.Cmd) error
}

type SystemAPIMock struct {
	StdOutText string
	StdErrText string
	Cmds       []*exec.Cmd
	ErrToGive  error

	UserNameInput      string
	UserNamePathInput  string
	UserSetErrToReturn error
	Callback           func(cmd *exec.Cmd) error
}

func (oem *SystemAPIMock) Run(cmd *exec.Cmd) error {
	oem.Cmds = append(oem.Cmds, cmd)

	if oem.Callback != nil {
		return oem.Callback(cmd)
	}

	if oem.StdOutText != "" && cmd.Stdout != nil {
		_, err := cmd.Stdout.Write([]byte(oem.StdOutText))
		if err != nil {
			return err
		}
	}

	if oem.StdErrText != "" && cmd.Stderr != nil {
		_, err := cmd.Stderr.Write([]byte(oem.StdErrText))
		if err != nil {
			return err
		}
	}

	return oem.ErrToGive
}

func (oem *SystemAPIMock) SetUser(userName, path string, cmd *exec.Cmd) error {
	oem.UserNameInput = userName
	oem.UserNamePathInput = path

	return oem.UserSetErrToReturn
}

type Runner interface {
	Run(execContext *Context) error
}

type SystemRunner struct {
	SystemAPI SystemAPI
}

type RunnerMock struct {
	GivenExecContexts []*Context
	ErrToReturn       error
	RunOutputCallback func(stdOutWriter, stdErrWriter io.Writer)
}

func (rm *RunnerMock) Run(execContext *Context) error {
	rm.GivenExecContexts = append(rm.GivenExecContexts, execContext)
	if rm.RunOutputCallback != nil {
		rm.RunOutputCallback(execContext.StdoutWriter, execContext.StderrWriter)
	}
	return rm.ErrToReturn
}

func (sr SystemRunner) Run(execContext *Context) error {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "taco-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	cmd, err := sr.createCmd(execContext, tmpFile)
	if err != nil {
		return err
	}

	var exitCode int
	execContext.Pid, exitCode, err = sr.runCmd(cmd)

	if err != nil {
		return RunError{Err: err, ExitCode: exitCode}
	}

	return nil
}

func (sr SystemRunner) runCmd(cmd *exec.Cmd) (pid, exitCode int, err error) {
	logrus.Debugf("will run cmd '%s'", cmd.String())
	err = sr.SystemAPI.Run(cmd)
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.ExitCode()
	}

	if err != nil {
		logrus.Debugf("stopped execution of multi-command")
		return pid, exitCode, err
	}
	logrus.Debugf("execution success for '%s'", cmd.String())

	return pid, exitCode, nil
}

func (sr SystemRunner) setWorkingDir(cmd *exec.Cmd, execContext *Context) {
	if execContext.WorkingDir != "" {
		logrus.Debugf("will set working dir %s to command %s", execContext.WorkingDir, cmd)
		cmd.Dir = execContext.WorkingDir
	}
}

func (sr SystemRunner) createCmd(execContext *Context, tmpFile *os.File) (cmd *exec.Cmd, err error) {
	rawCmds := strings.Join(execContext.Cmds, "\n")
	if _, err := tmpFile.Write([]byte(rawCmds)); err != nil {
		return nil, err
	}

	shellParam := sr.parseShellParam(execContext.Shell)
	cmdName := sr.buildCmdParts(shellParam)

	cmd = exec.Command(cmdName, tmpFile.Name())

	sr.setWorkingDir(cmd, execContext)
	if err := sr.setUser(cmd, execContext); err != nil {
		return nil, err
	}

	sr.setEnvs(cmd, execContext)
	sr.setIO(cmd, execContext.StdoutWriter, execContext.StderrWriter)
	return
}

func (sr SystemRunner) setEnvs(cmd *exec.Cmd, execContext *Context) {
	if len(execContext.Envs) == 0 {
		return
	}

	envs := execContext.Envs.ToEqualSignStrings()
	logrus.Debugf("will set %d env variables: %s to command '%s'", len(envs), envs, cmd)
	cmd.Env = append(os.Environ(), envs...)
}

func (sr SystemRunner) setUser(cmd *exec.Cmd, execContext *Context) error {
	if execContext.User == "" {
		return nil
	}
	err := sr.SystemAPI.SetUser(execContext.User, execContext.Path, cmd)

	if err != nil {
		return err
	}

	return nil
}

func (sr SystemRunner) buildCmdParts(shellParam ShellParam) (cmdName string) {
	cmdName = shellParam.ShellPath
	return
}

func (sr SystemRunner) setIO(cmd *exec.Cmd, stdOutWriter, stdErrWriter io.Writer) {
	logrus.Debugf("will set stdout and stderr to cmd '%s'", cmd)
	stdOutLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Debugf("stdout capture: %s", string(p))
			return len(p), nil
		},
	}
	stdErrLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Debugf("stderr capture: %s", string(p))
			return len(p), nil
		},
	}
	cmd.Stdout = io.MultiWriter(stdOutLoggedWriter, stdOutWriter)
	cmd.Stderr = io.MultiWriter(stdErrLoggedWriter, stdErrWriter)
}

func (sr SystemRunner) parseShellParam(rawShell string) ShellParam {
	rawShell = strings.TrimSpace(rawShell)
	if rawShell == "" {
		if runtime.GOOS == "windows" {
			rawShell = defaultWindowsShell
		} else {
			rawShell = defaultUnixShell
		}
	}

	parsedShellParam := ShellParam{
		RawShellString: rawShell,
	}

	shellParts := strings.Split(rawShell, " ")
	parsedShellParam.ShellPath = shellParts[0]

	shellPathParts := strings.Split(parsedShellParam.ShellPath, string(os.PathSeparator))
	parsedShellParam.ShellName = shellPathParts[len(shellPathParts)-1]

	return parsedShellParam
}
