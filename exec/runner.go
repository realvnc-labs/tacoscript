package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	io2 "github.com/cloudradar-monitoring/tacoscript/io"
	"github.com/sirupsen/logrus"
)

var cParamShells = map[string]string{
	"zsh":     "-c",
	"bash":    "-c",
	"sh":      "-c",
	"csh":     "-c",
	"tcsh":    "-c",
	"ksh":     "-c",
	"fish":    "-c",
	"cmd.exe": "/C",
}

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
	ErrToReturn error
	RunOutputCallback func(stdOutWriter, stdErrWriter io.Writer)
}

func (rm *RunnerMock) Run(execContext *Context) error{
	rm.GivenExecContexts = append(rm.GivenExecContexts, execContext)
	if rm.RunOutputCallback != nil {
		rm.RunOutputCallback(execContext.StdoutWriter, execContext.StderrWriter)
	}
	return rm.ErrToReturn
}

func (sr SystemRunner) Run(execContext *Context) error {
	cmds, err := sr.createCmds(execContext)
	if err != nil {
		return err
	}

	err = sr.runCmds(cmds)
	if err != nil {
		return RunError{Err: err}
	}

	return nil
}

func (sr SystemRunner) runCmds(cmds []*exec.Cmd) error {
	for _, cmd := range cmds {
		logrus.Debugf("will run cmd '%s'", cmd.String())
		err := sr.SystemAPI.Run(cmd)
		if err != nil {
			return err
		}
		logrus.Debugf("execution success for '%s'", cmd.String())
	}

	return nil
}

func (sr SystemRunner) setWorkingDir(cmd *exec.Cmd, execContext *Context) {
	if execContext.WorkingDir != "" {
		logrus.Debugf("will set working dir %s to command %s", execContext.WorkingDir, cmd)
		cmd.Dir = execContext.WorkingDir
	}
}

func (sr SystemRunner) createCmds(execContext *Context) (cmds []*exec.Cmd, err error) {
	rawCmds := make([]string, 0, len(execContext.Cmds))
	for _, cmdName := range execContext.Cmds {
		cmdName = strings.TrimSpace(cmdName)
		if cmdName == "" {
			continue
		}

		rawCmds = append(rawCmds, cmdName)
	}

	shellParam := sr.parseShellParam(execContext.Shell)
	for _, rawCmd := range rawCmds {
		cmd, err := sr.createCmd(rawCmd, shellParam, execContext)
		if err != nil {
			return cmds, err
		}

		cmds = append(cmds, cmd)
	}

	return
}

func (sr SystemRunner) createCmd(rawCmd string, shellParam ShellParam, execContext *Context) (*exec.Cmd, error) {
	cmdParam := sr.parseCmdParam(rawCmd)

	cmdName, cmdArgs := sr.buildCmdParts(shellParam, cmdParam)

	cmd := exec.Command(cmdName, cmdArgs...)

	sr.setWorkingDir(cmd, execContext)

	err := sr.setUser(cmd, execContext)
	if err != nil {
		return nil, err
	}

	sr.setEnvs(cmd, execContext)
	sr.setIO(cmd, execContext.StdoutWriter, execContext.StderrWriter)

	return cmd, nil
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

func (sr SystemRunner) parseCmdParam(rawCmd string) CmdParam {
	rawCmd = strings.TrimSpace(rawCmd)

	parsedCmdParam := CmdParam{
		RawCmdString: rawCmd,
	}

	if rawCmd == "" {
		return parsedCmdParam
	}

	cmdParts := strings.Split(rawCmd, " ")
	parsedCmdParam.Cmd = cmdParts[0]

	parsedCmdParam.Params = make([]string, 0, len(cmdParts))
	for k, cmdPart := range cmdParts {
		if k == 0 {
			continue
		}
		cmdPart = strings.TrimSpace(cmdPart)
		parsedCmdParam.Params = append(parsedCmdParam.Params, cmdPart)
	}

	return parsedCmdParam
}

func (sr SystemRunner) buildCmdParts(shellParam ShellParam, cmdParam CmdParam) (cmdName string, cmdArgs []string) {
	if shellParam.ShellPath != "" {
		sr.addCShellParamIfNeeded(&shellParam)
		cmdName = shellParam.ShellPath
		cmdParams := fmt.Sprintf("%s %s", cmdParam.Cmd, strings.Join(cmdParam.Params, " "))
		cmdArgs = shellParam.ShellParams
		cmdArgs = append(cmdArgs, cmdParams)
	} else {
		cmdName = cmdParam.Cmd
		cmdArgs = cmdParam.Params
	}

	return
}

func (sr SystemRunner) addCShellParamIfNeeded(shellParam *ShellParam) {
	if shellParam.ShellName == "" || len(shellParam.ShellParams) > 0 {
		return
	}

	var cParam string
	for knownShellName, knownCParam := range cParamShells {
		if knownShellName == shellParam.ShellName {
			cParam = knownCParam
			break
		}
	}

	if cParam != "" {
		shellParam.ShellParams = append(shellParam.ShellParams, cParam)
	}
}

func (sr SystemRunner) setIO(cmd *exec.Cmd, stdOutWriter, stdErrWriter io.Writer) {
	logrus.Debugf("will set stdout and stderr to cmd '%s'", cmd)
	stdOutLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Infof(string(p))
			return len(p), nil
		},
	}
	stdErrLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Warn(string(p))
			return len(p), nil
		},
	}
	cmd.Stdout = io.MultiWriter(stdOutLoggedWriter, stdOutWriter)
	cmd.Stderr = io.MultiWriter(stdErrLoggedWriter, stdErrWriter)
}

func (sr SystemRunner) parseShellParam(rawShell string) ShellParam {
	rawShell = strings.TrimSpace(rawShell)

	parsedShellParam := ShellParam{
		RawShellString: rawShell,
	}

	if rawShell == "" {
		return parsedShellParam
	}

	shellParts := strings.Split(rawShell, " ")
	parsedShellParam.ShellPath = shellParts[0]

	parsedShellParam.ShellParams = make([]string, 0, len(shellParts))
	for k, shellPart := range shellParts {
		if k == 0 {
			continue
		}
		shellPart = strings.TrimSpace(shellPart)
		parsedShellParam.ShellParams = append(parsedShellParam.ShellParams, shellPart)
	}

	shellPathParts := strings.Split(parsedShellParam.ShellPath, string(os.PathSeparator))
	parsedShellParam.ShellName = shellPathParts[len(shellPathParts)-1]

	return parsedShellParam
}
