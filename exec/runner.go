package exec

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	io2 "github.com/realvnc-labs/tacoscript/io"
	"github.com/sirupsen/logrus"
)

// the default windows shell must be cmd.exe for compatibility with older Windows versions
const defaultWindowsShell = "cmd.exe"

const defaultUnixShell = "sh"

var powershellShells = []string{"powershell", "powershell.exe", "pwsh", "pwsh.exe"}

func IsPowerShell(shell string) bool {
	for _, name := range powershellShells {
		if name == shell {
			return true
		}
	}
	return false
}

type SystemAPI interface {
	Run(cmd *exec.Cmd) error
	SetUser(userName, path string, cmd *exec.Cmd) error
}

type Runner interface {
	Run(execContext *Context) error
}

type SystemRunner struct {
	SystemAPI SystemAPI
}

func (sr SystemRunner) Run(execContext *Context) error {
	if execContext.Shell == "" {
		if runtime.GOOS == "windows" {
			execContext.Shell = defaultWindowsShell
		} else {
			execContext.Shell = defaultUnixShell
		}
	}
	if execContext.Shell == "cmd" && runtime.GOOS == "windows" {
		execContext.Shell = defaultWindowsShell
	}

	tmpPattern := "taco-*"
	if runtime.GOOS == "windows" {
		if execContext.Shell == defaultWindowsShell {
			tmpPattern = "taco-*.cmd"
		} else {
			tmpPattern = "taco-*.ps1"
		}
	}
	tmpFile, err := os.CreateTemp(os.TempDir(), tmpPattern)
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

	logrus.Debugf("execution success for '%s'", cmd.String())
	return pid, exitCode, err
}

func (sr SystemRunner) setWorkingDir(cmd *exec.Cmd, execContext *Context) {
	if execContext.WorkingDir != "" {
		logrus.Debugf("will set working dir %s to command %s", execContext.WorkingDir, cmd)
		cmd.Dir = execContext.WorkingDir
	}
}

func (sr SystemRunner) createCmd(execContext *Context, tmpFile *os.File) (cmd *exec.Cmd, err error) {
	prelude := ""
	newLine := "\n"

	if runtime.GOOS == "windows" {
		newLine = "\r\n"
		if execContext.Shell == defaultWindowsShell {
			prelude = "@echo off" + newLine
		}
	}

	rawCmds := prelude + strings.Join(execContext.Cmds, newLine)

	if _, err = tmpFile.WriteString(rawCmds); err != nil {
		return
	}

	tmpFile.Close()

	logrus.Debugf("WROTE TO FILE:\n%s\n----\n", rawCmds)

	shellParam := sr.parseShellParam(execContext.Shell)
	cmdName, cmdArgs := sr.buildCmdParts(shellParam)

	if runtime.GOOS == "windows" && execContext.Shell == defaultWindowsShell {
		cmdName = tmpFile.Name()
	} else {
		cmdArgs = append(cmdArgs, tmpFile.Name())
	}

	cmd = exec.Command(cmdName, cmdArgs...)

	sr.setWorkingDir(cmd, execContext)
	if err = sr.setUser(cmd, execContext); err != nil {
		return
	}

	sr.setEnvs(cmd, execContext)
	sr.setIO(cmd, execContext.StdoutWriter, execContext.StderrWriter)
	return cmd, err
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

func (sr SystemRunner) buildCmdParts(shellParam ShellParam) (cmdName string, cmdArgs []string) {
	cmdName = shellParam.ShellPath
	cmdArgs = shellParam.ShellParams
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
	parsedShellParam := ShellParam{
		RawShellString: rawShell,
	}

	shellParts := strings.Split(rawShell, " ")
	parsedShellParam.ShellPath = shellParts[0]

	shellPathParts := strings.Split(parsedShellParam.ShellPath, string(os.PathSeparator))
	parsedShellParam.ShellName = shellPathParts[len(shellPathParts)-1]

	for k, shellPart := range shellParts {
		if k == 0 {
			continue
		}
		shellPart = strings.TrimSpace(shellPart)
		parsedShellParam.ShellParams = append(parsedShellParam.ShellParams, shellPart)
	}
	return parsedShellParam
}
