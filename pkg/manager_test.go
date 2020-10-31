package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

type MockedPackageManagerCmdBuilder struct {
}

func (mpmb *MockedPackageManagerCmdBuilder) GetPkgManagerVersionCmd() string {
	return "mpmb --version"
}

func (mpmb *MockedPackageManagerCmdBuilder) GetUpdateCmd(t *tasks.PkgTask) string {
	return "mpmb upgrade"
}

func (mpmb *MockedPackageManagerCmdBuilder) GetInstallCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()

	versionStr := ""
	if t.Version != ""{
		versionStr = "--version " + t.Version + " "
	}

	return []string{fmt.Sprintf("mpmb install %s%s", versionStr, strings.Join(rawCmds, " "))}
}

func (mpmb *MockedPackageManagerCmdBuilder) GetUninstallCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()

	return []string{fmt.Sprintf("mpmb uninstall %s", strings.Join(rawCmds, " "))}
}

func (mpmb *MockedPackageManagerCmdBuilder) GetUpdateCmdForPackages(t *tasks.PkgTask) []string {
	rawCmds := t.GetNames()

	return []string{fmt.Sprintf("mpmb update %s", strings.Join(rawCmds, " "))}
}

func TestTaskExecution(t *testing.T) {
	testCases := []struct {
		Runner *exec.RunnerMock
		Task *tasks.PkgTask
		Name string
		ExpectedCmds []string
		ExpectedOutput string
		ExpectedErrStr string
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
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionInstall,
				NamedTask: tasks.NamedTask{Name: "vim"},
				Path: "somePath",
				Shell: "sh",
			},
			ExpectedOutput: "some stdout",
			ExpectedCmds: []string{"mpmb --version", "mpmb install vim"},
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
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionInstall,
				NamedTask: tasks.NamedTask{Name: "vim"},
				Version: "1.1.0",
			},
			ExpectedOutput: "some stderr",
			ExpectedCmds: []string{"mpmb --version", "mpmb install --version 1.1.0 vim"},
		},
		{
			Name: "multiple_name_install_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionInstall,
				NamedTask: tasks.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb install vim nano"},
		},
		{
			Name: "multiple_name_uninstall_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionUninstall,
				NamedTask: tasks.NamedTask{Names: []string{"vim", "nano"}, Name: "mc"},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb uninstall mc vim nano"},
		},
		{
			Name: "multiple_name_update_success",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionUpdate,
				NamedTask: tasks.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb update vim nano"},
		},
		{
			Name: "non_existing_pkg_manager",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
				ErrToReturn: errors.New("non existing pkg manager"),
			},
			Task: &tasks.PkgTask{
				ActionType: tasks.ActionInstall,
				NamedTask: tasks.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version"},
			ExpectedErrStr: "non existing pkg manager",
		},
		{
			Name: "pkg_manager_update_before_install",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &tasks.PkgTask{
				ShouldRefresh: true,
				ActionType: tasks.ActionUpdate,
				NamedTask: tasks.NamedTask{Names: []string{"vim", "nano"}},
			},
			ExpectedCmds: []string{"mpmb --version", "mpmb upgrade", "mpmb update vim nano"},
		},
		{
			Name: "invalid_pkg_action_type",
			Runner: &exec.RunnerMock{
				GivenExecContexts: []*exec.Context{},
			},
			Task: &tasks.PkgTask{
				TypeName: "some uknown type name",
				ActionType: 0,
			},
			ExpectedErrStr: "unknown action type '0' for task some uknown type name",
			ExpectedCmds: []string{"mpmb --version"},
		},
	}

	for _, testCase := range testCases{
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			mngr := PackageTaskManager{
				Runner: tc.Runner,
				PackageManagerCmdBuilder: &MockedPackageManagerCmdBuilder{},
			}

			output, err := mngr.ExecuteTask(context.Background(), tc.Task)

			assert.Equal(t, tc.ExpectedOutput, output)

			if tc.ExpectedErrStr != "" {
				assert.EqualError(t, err, tc.ExpectedErrStr)
			} else {
				assert.NoError(t, err)
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
