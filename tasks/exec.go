package tasks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	io2 "github.com/cloudradar-monitoring/tacoscript/io"
	"github.com/sirupsen/logrus"
)

type CmdRunner interface {
	Run(cmd *exec.Cmd) error
}

type OSCmdRunner struct{}

func (ocm OSCmdRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

type UserSystemInfoParser interface {
	Parse(userName, path string) (sysUserID, sysGroupID uint32, err error)
}

type OSUserSystemInfoParser struct{}

func (ousif OSUserSystemInfoParser) Parse(userName, path string) (sysUserID, sysGroupID uint32, err error) {
	logrus.Debugf("parsing user '%s' to get uid and group id from OS", userName)
	u, err := user.Lookup(userName)
	if err != nil {
		err = fmt.Errorf("cannot locate user '%s': %w, check path '%s'", userName, err, path+"."+UserField)
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		err = fmt.Errorf("non-numeric user ID '%s': %w, check path '%s'", u.Uid, err, path+"."+UserField)
		return
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		err = fmt.Errorf("non-numeric user group ID '%s': %w, check path '%s'", u.Gid, err, path+"."+UserField)
		return
	}

	return uint32(uid), uint32(gid), nil
}

type CmdRunTask struct {
	TypeName             string
	Path                 string
	Cmd                  string
	WorkingDir           string
	User                 string
	Shell                string
	Envs                 Envs
	MissingFileCondition string
	Errors               *ValidationErrors
	Runner               CmdRunner
	UserSystemInfoParser UserSystemInfoParser
}

type CmdRunTaskBuilder struct {
	Runner               CmdRunner
	UserSystemInfoParser UserSystemInfoParser
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, ctx []map[string]interface{}) (Task, error) {
	t := &CmdRunTask{
		TypeName:             typeName,
		Path:                 path,
		Errors:               &ValidationErrors{},
		Runner:               crtb.Runner,
		UserSystemInfoParser: crtb.UserSystemInfoParser,
	}

	for _, contextItem := range ctx {
		for key, val := range contextItem {
			switch key {
			case NameField:
				t.Cmd = fmt.Sprint(val)
			case CwdField:
				t.WorkingDir = fmt.Sprint(val)
			case UserField:
				t.User = fmt.Sprint(val)
			case ShellField:
				t.Shell = fmt.Sprint(val)
			case EnvField:
				envs, err := extractEnvFields(val, path)
				t.Errors.Add(err)
				t.Envs = envs
			case CreatesField:
				t.MissingFileCondition = fmt.Sprint(val)
			}
		}
	}

	return t, nil
}

func (crt *CmdRunTask) GetName() string {
	return crt.TypeName
}

func (crt *CmdRunTask) Validate() error {
	crt.Errors.Add(ValidateRequired(crt.Cmd, crt.Path+"."+CwdField))
	return crt.Errors.ToError()
}

func (crt *CmdRunTask) GetPath() string {
	return crt.Path
}

func (crt *CmdRunTask) Execute(ctx context.Context) ExecutionResult {
	execRes := ExecutionResult{}

	execRes.IsSkipped, execRes.Err = crt.checkMissingFileCondition()
	if execRes.Err != nil {
		return execRes
	}
	if execRes.IsSkipped {
		return execRes
	}

	cmd := crt.createCmd()
	crt.setWorkingDir(cmd)

	execRes.Err = crt.setUser(cmd)
	if execRes.Err != nil {
		return execRes
	}

	crt.setEnvs(cmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	crt.setIO(cmd, &stdoutBuf, &stderrBuf)

	crt.run(cmd, &execRes)

	execRes.StdErr = stderrBuf.String()
	execRes.StdOut = stdoutBuf.String()

	return execRes
}

func (crt *CmdRunTask) checkMissingFileCondition() (isSkipped bool, err error) {
	if crt.MissingFileCondition == "" {
		return
	}

	logrus.Debugf("will check if file '%s' is missing", crt.MissingFileCondition)
	_, err = os.Stat(crt.MissingFileCondition)
	if err == nil {
		logrus.Infof("file %s exists, will skip command '%s'", crt.MissingFileCondition, crt.Cmd)
		isSkipped = true
		return
	}

	if !os.IsNotExist(err) {
		err = fmt.Errorf("failed to check if file '%s' exists: %w", crt.MissingFileCondition, err)
		return
	}

	return
}

func (crt *CmdRunTask) createCmd() *exec.Cmd {
	shellParam := crt.parseShellParam(crt.Shell)

	cmdParam := crt.parseCmdParam(crt.Cmd)

	cmdName, cmdArgs := crt.buildCmdParts(shellParam, cmdParam)
	cmd := exec.Command(cmdName, cmdArgs...)

	return cmd
}

func (crt *CmdRunTask) buildCmdParts(shellParam ShellParam, cmdParam CmdParam) (cmdName string, cmdArgs []string) {
	if shellParam.ShellPath != "" {
		crt.addCShellParamIfNeeded(&shellParam)
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

func (crt *CmdRunTask) addCShellParamIfNeeded(shellParam *ShellParam) {
	if shellParam.ShellName == "" || len(shellParam.ShellParams) > 0 {
		return
	}

	isCParamSupported := false
	for _, cShell := range cParamShells {
		if cShell == shellParam.ShellName {
			isCParamSupported = true
			break
		}
	}

	if isCParamSupported {
		shellParam.ShellParams = append(shellParam.ShellParams, "-c")
	}
}

func (crt *CmdRunTask) setWorkingDir(cmd *exec.Cmd) {
	if crt.WorkingDir != "" {
		logrus.Debugf("will set working dir %s", crt.WorkingDir)
		cmd.Dir = crt.WorkingDir
	}
}

func (crt *CmdRunTask) setUser(cmd *exec.Cmd) error {
	if crt.User == "" {
		return nil
	}
	logrus.Debugf("will set user %s", crt.User)

	uid, gid, err := crt.UserSystemInfoParser.Parse(crt.User, crt.Path)
	if err != nil {
		return err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Credential: &syscall.Credential{Uid: uid, Gid: gid},
	}

	return nil
}

func (crt *CmdRunTask) setEnvs(cmd *exec.Cmd) {
	if len(crt.Envs) == 0 {
		return
	}

	envs := crt.Envs.ToOSRaw()
	logrus.Debugf("will set %d env variables", len(envs))
	cmd.Env = append(os.Environ(), envs...)
}

func (crt *CmdRunTask) setIO(cmd *exec.Cmd, stdOutWriter, stdErrWriter io.Writer) {
	stdOutLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Infof(string(p))
			return len(p), nil
		},
	}
	stdErrLoggedWriter := io2.FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			logrus.Errorf(string(p))
			return len(p), nil
		},
	}
	cmd.Stdout = io.MultiWriter(stdOutLoggedWriter, stdOutWriter)
	cmd.Stderr = io.MultiWriter(stdErrLoggedWriter, stdErrWriter)
}

func (crt *CmdRunTask) run(cmd *exec.Cmd, res *ExecutionResult) {
	logrus.Infof("will run cmd %s", cmd.String())

	start := time.Now()

	err := crt.Runner.Run(cmd)

	res.Duration = time.Since(start)
	if err != nil {
		res.Err = err
	}

	logrus.Debugf("execution of %s has finished, took: %v", crt.Cmd, res.Duration)
}

func (crt *CmdRunTask) parseShellParam(rawShell string) ShellParam {
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

func (crt *CmdRunTask) parseCmdParam(rawCmd string) CmdParam {
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
