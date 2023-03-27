package pkgmanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks/namedtask"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
	"github.com/stretchr/testify/assert"
)

type MockedOsPackageManagerCmdProvider struct {
	ErrToGive error
}

func (ecb MockedOsPackageManagerCmdProvider) GetManagementCmds(t *pkgtask.PTask) (*ManagementCmds, error) {
	rawCmds := t.Named.GetNames()

	versionStr := ""
	if t.Version != "" {
		versionStr = "--version " + t.Version + " "
	}

	return &ManagementCmds{
		VersionCmd:    "mpmb --version",
		UpgradeCmd:    "mpmb upgrade",
		InstallCmds:   []string{fmt.Sprintf("mpmb install %s%s", versionStr, strings.Join(rawCmds, " "))},
		UninstallCmds: []string{fmt.Sprintf("mpmb uninstall %s", strings.Join(rawCmds, " "))},
		UpgradeCmds:   []string{fmt.Sprintf("mpmb update %s", strings.Join(rawCmds, " "))},
		ListCmd:       "mpmb list",
	}, ecb.ErrToGive
}

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		Runner           *exec.RunnerMock
		Task             *pkgtask.PTask
		Name             string
		CmdBuildErrorStr string
		ExpectedCmds     []string
		ExpectedOutput   string
		ExpectedErrStr   string
	}{
		{
			Name: "single_name_install_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
				RunOutputCallback: func(stdOutWriter, stdErrWriter io.Writer) {
					_, err := stdOutWriter.Write([]byte("some stdout"))
					assert.NoError(t, err)
				},
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionInstall,
				Named:      namedtask.NamedTask{Name: "vim"},
				Path:       "somePath",
				Shell:      "sh",
			},
			ExpectedOutput: "some stdout",
			ExpectedCmds:   []string{"mpmb --version", "mpmb list", "mpmb install vim", "mpmb list"},
		},
		{
			Name: "single_name_install_version_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
				RunOutputCallback: func(stdOutWriter, stdErrWriter io.Writer) {
					_, err := stdErrWriter.Write([]byte("some stderr"))
					assert.NoError(t, err)
				},
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionInstall,
				Named:      namedtask.NamedTask{Name: "vim"},
				Version:    "1.1.0",
			},
			ExpectedOutput: "some stderr",
			ExpectedCmds:   []string{"mpmb --version", "mpmb list", "mpmb install --version 1.1.0 vim", "mpmb list"},
		},
		{
			Name: "multiple_name_install_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionInstall,
				Named:      namedtask.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb list", "mpmb install vim nano", "mpmb list"},
		},
		{
			Name: "multiple_name_uninstall_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionUninstall,
				Named:      namedtask.NamedTask{Names: []string{"vim", "nano"}, Name: "mc"},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb list", "mpmb uninstall mc vim nano", "mpmb list"},
		},
		{
			Name: "multiple_name_update_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionUpdate,
				Named:      namedtask.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb list", "mpmb update vim nano", "mpmb list"},
		},
		{
			Name: "non_existing_pkg_manager",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
				ErrToReturn:       errors.New("non existing pkg manager"),
			},
			Task: &pkgtask.PTask{
				ActionType: pkgtask.ActionInstall,
				Named:      namedtask.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds:   []string{"mpmb --version"},
			ExpectedErrStr: "cannot find a supported package manager on the host, tried package manager commands: mpmb --version: non existing pkg manager",
		},
		{
			Name: "pkg_manager_update_before_install",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				ShouldRefresh: true,
				ActionType:    pkgtask.ActionUpdate,
				Named:         namedtask.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb upgrade", "mpmb list", "mpmb update vim nano", "mpmb list"},
		},
		{
			Name: "invalid_pkg_action_type",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				TypeName:   "some unknown type name",
				ActionType: 0,
			},
			ExpectedErrStr: "unknown action type '0' for task some unknown type name",
			ExpectedCmds:   []string{"mpmb --version", "mpmb list"},
		},
		{
			Name: "build_cmd_error",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &pkgtask.PTask{
				TypeName:   "some uknown type name",
				ActionType: pkgtask.ActionInstall,
			},
			CmdBuildErrorStr: "cannot build command",
			ExpectedErrStr:   "cannot build command",
			ExpectedCmds:     []string{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			var cmdBuildError error
			if tc.CmdBuildErrorStr != "" {
				cmdBuildError = errors.New(tc.CmdBuildErrorStr)
			}
			mngr := PackageTaskManager{
				Runner: tc.Runner,
				ManagementCmdsProviderBuildFunc: func() ([]ManagementCmdsProvider, error) {
					return []ManagementCmdsProvider{
						&MockedOsPackageManagerCmdProvider{ErrToGive: cmdBuildError},
					}, nil
				},
			}

			res, err := mngr.ExecuteTask(context.Background(), tc.Task)

			if tc.ExpectedErrStr != "" {
				assert.EqualError(t, err, tc.ExpectedErrStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedOutput, res.Output)
			}

			actualCmds := make([]string, 0, len(tc.Runner.GivenExecContexts))
			for _, execContext := range tc.Runner.GivenExecContexts {
				assert.Equal(t, tc.Task.Path, execContext.Path)
				assert.Equal(t, tc.Task.Shell, execContext.Shell)
				actualCmds = append(actualCmds, execContext.Cmds...)
			}
			assert.Equal(t, tc.ExpectedCmds, actualCmds)
		})
	}
}
