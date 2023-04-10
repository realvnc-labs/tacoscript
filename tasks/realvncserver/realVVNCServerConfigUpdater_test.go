//go:build !windows
// +build !windows

package realvncserver

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/tasks/fieldstatus"
	"github.com/realvnc-labs/tacoscript/utils"
)

const (
	origTestConfigFilename = "../../realvnc/test/realvncserver-config.conf.orig"
	testConfigFilename     = "../../realvnc/test/realvncserver-config.conf"
)

func testSetup(t *testing.T) {
	t.Helper()
	makeTestConfigFile(t)
}

func testTeardown(t *testing.T) {
	t.Helper()
	removeTestConfigFile(t)
}

func makeTestConfigFile(t *testing.T) {
	contents, err := os.ReadFile(origTestConfigFilename)
	require.NoError(t, err)
	err = os.WriteFile(testConfigFilename, contents, 0644) //nolint:gosec // test file
	require.NoError(t, err)
}

func removeTestConfigFile(t *testing.T) {
	err := os.Remove(testConfigFilename)
	require.NoError(t, err)

	_ = os.Remove(utils.GetBackupFilename(testConfigFilename, "bak"))
}

func TestShouldUpdateSimpleConfigFileParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap("encryption", "Encryption"),
		fieldstatus.StatusMap{
			"Encryption": fieldstatus.FieldStatus{
				HasNewValue: true,
			},
		})

	task := &RvsTask{
		Path:         "realvnc-server-1",
		ConfigFile:   "../../realvnc/test/realvncserver-config.conf",
		Encryption:   "AlwaysOn",
		fieldMapper:  tracker,
		fieldTracker: tracker,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "Config updated")
	assert.Equal(t, res.Changes["count"], "1 config value change(s) applied")

	contents, err := os.ReadFile(task.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "Encryption=AlwaysOn")

	backupContents, err := os.ReadFile(utils.GetBackupFilename(task.ConfigFile, "bak"))
	require.NoError(t, err)
	assert.Contains(t, string(backupContents), "Encryption=BadValue")
}

func TestShouldAddSimpleConfigFileParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap("blank_screen", "BlankScreen"),
		fieldstatus.StatusMap{
			"BlankScreen": fieldstatus.FieldStatus{
				HasNewValue: true,
			},
		})

	task := &RvsTask{
		Path:        "realvnc-server-1",
		ConfigFile:  "../../realvnc/test/realvncserver-config.conf",
		SkipBackup:  true,
		BlankScreen: true,
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "Config updated", res.Comment)
	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	contents, err := os.ReadFile(task.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "BlankScreen=true")

	_, err = os.ReadFile(utils.GetBackupFilename(task.ConfigFile, "bak"))
	require.ErrorContains(t, err, "no such file")
}

func TestShouldAddSimpleConfigWhenNoExistingConfigFile(t *testing.T) {
	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	fmt.Printf("cwd = %+v\n", cwd)

	newConfigFilename := "../../realvnc/test/realvncserver-config-new.conf"
	defer os.Remove(newConfigFilename)

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap("idle_timeout", "IdleTimeout"),
		fieldstatus.StatusMap{
			"IdleTimeout": fieldstatus.FieldStatus{
				HasNewValue: true,
			},
		})

	task := &RvsTask{
		Path:        "realvnc-server-1",
		ConfigFile:  newConfigFilename,
		IdleTimeout: 3600,
		SkipBackup:  false,
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "Config updated", res.Comment)
	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	contents, err := os.ReadFile(task.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "IdleTimeout=3600")

	info, err := os.Stat(task.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, fs.FileMode(DefaultConfigFilePermissions), info.Mode().Perm())

	_, err = os.ReadFile(utils.GetBackupFilename(task.ConfigFile, "bak"))
	require.ErrorContains(t, err, "no such file")
}

func TestShouldRemoveSimpleConfigFileParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap("encryption", "Encryption"),
		fieldstatus.StatusMap{
			"Encryption": fieldstatus.FieldStatus{
				HasNewValue: true,
				Clear:       true,
			},
		})

	task := &RvsTask{
		Path:       "realvnc-server-1",
		ConfigFile: "../../realvnc/test/realvncserver-config.conf",
		Encryption: "!UNSET!",
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "Config updated", res.Comment)
	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	contents, err := os.ReadFile(task.ConfigFile)
	require.NoError(t, err)

	assert.NotContains(t, string(contents), "Encryption")
}

func TestShouldMakeReloadCmdLine(t *testing.T) {
	cases := []struct {
		name            string
		task            RvsTask
		goos            string
		expectedCmdLine string
		expectedParams  []string
	}{
		{
			name: "linux service server mode",
			task: RvsTask{
				Path:       "MyTask",
				ServerMode: ServiceServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vncserver-x11",
			expectedParams:  []string{"-service", "-reload"},
		},
		{
			name: "linux user server mode",
			task: RvsTask{
				Path:       "MyTask",
				ServerMode: UserServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vncserver-x11",
			expectedParams:  []string{"-reload"},
		},
		{
			name: "linux virtual server mode",
			task: RvsTask{
				Path:       "MyTask",
				ServerMode: VirtualServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vnclicense",
			expectedParams:  []string{"-reload"},
		},
		{
			name: "user specified",
			task: RvsTask{
				Path:           "MyTask",
				ServerMode:     ServiceServerMode,
				ReloadExecPath: "/my/path/vncserver-x11",
			},
			goos:            "linux",
			expectedCmdLine: "/my/path/vncserver-x11",
			expectedParams:  []string{"-service", "-reload"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			//nolint: gocritic // code will be re-enabed in the future. see comment below.
			// err := task.Validate(tc.goos)
			// require.NoError(t, err)

			// TODO: (rs): no need to set the flag once validations are re-introduced with the additional
			// server modes.

			if task.ServerMode == VirtualServerMode {
				task.UseVNCLicenseReload = true
			}

			cmdLine, params := makeReloadCmdLine(&task, tc.goos)
			assert.Equal(t, tc.expectedCmdLine, cmdLine)
			assert.Equal(t, tc.expectedParams, params)
		})
	}
}
