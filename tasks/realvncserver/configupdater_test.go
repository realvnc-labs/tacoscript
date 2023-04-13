//go:build !windows
// +build !windows

package realvncserver_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver/rvstbuilder"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/utils"
)

const (
	origTestConfigFilename = "../../testdata/realvncserver-config.conf.orig"
	testConfigFilename     = "../../testdata/realvncserver-config.conf"
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

	executor := &realvncserver.RvstExecutor{
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

	task := &realvncserver.RvsTask{
		Path:       "realvnc-server-1",
		ConfigFile: testConfigFilename,
		Encryption: "AlwaysOn",
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

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

	executor := &realvncserver.RvstExecutor{
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

	task := &realvncserver.RvsTask{
		Path:        "realvnc-server-1",
		ConfigFile:  testConfigFilename,
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

	executor := &realvncserver.RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	fmt.Printf("cwd = %+v\n", cwd)

	newConfigFilename := "../../testdata/realvncserver-config-new.conf"
	defer os.Remove(newConfigFilename)

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap("idle_timeout", "IdleTimeout"),
		fieldstatus.StatusMap{
			"IdleTimeout": fieldstatus.FieldStatus{
				HasNewValue: true,
			},
		})

	task := &realvncserver.RvsTask{
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
	assert.Equal(t, fs.FileMode(realvncserver.DefaultConfigFilePermissions), info.Mode().Perm())

	_, err = os.ReadFile(utils.GetBackupFilename(task.ConfigFile, "bak"))
	require.ErrorContains(t, err, "no such file")
}

func TestShouldRemoveSimpleConfigFileParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &realvncserver.RvstExecutor{
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

	task := &realvncserver.RvsTask{
		Path:       "realvnc-server-1",
		ConfigFile: testConfigFilename,
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
		task            realvncserver.RvsTask
		goos            string
		expectedCmdLine string
		expectedParams  []string
	}{
		{
			name: "linux service server mode",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.ServiceServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vncserver-x11",
			expectedParams:  []string{"-service", "-reload"},
		},
		{
			name: "linux user server mode",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.UserServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vncserver-x11",
			expectedParams:  []string{"-reload"},
		},
		{
			name: "linux virtual server mode",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.VirtualServerMode,
			},
			goos:            "linux",
			expectedCmdLine: "/usr/bin/vnclicense",
			expectedParams:  []string{"-reload"},
		},
		{
			name: "user specified",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ServerMode:     realvncserver.ServiceServerMode,
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

			if task.ServerMode == realvncserver.VirtualServerMode {
				task.UseVNCLicenseReload = true
			}

			cmdLine, params := realvncserver.MakeReloadCmdLine(&task, tc.goos)
			assert.Equal(t, tc.expectedCmdLine, cmdLine)
			assert.Equal(t, tc.expectedParams, params)
		})
	}
}

const targetConfigFilename = "./testconfig.conf"

func TestShouldHandleBackups(t *testing.T) {
	cases := []struct {
		name               string
		values             []interface{}
		sourceConfigFile   string
		targetConfigFile   string
		expectedBackupFile string
	}{
		{
			name:             "regular backup",
			sourceConfigFile: origTestConfigFilename,
			targetConfigFile: targetConfigFilename,
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: targetConfigFilename}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipReloadField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.EncryptionField, Value: "AlwaysOn"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BlankScreenField, Value: "true"}},
			},
			expectedBackupFile: targetConfigFilename + ".bak",
		},
		{
			name:             "alt backup",
			sourceConfigFile: origTestConfigFilename,
			targetConfigFile: targetConfigFilename,
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: targetConfigFilename}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipReloadField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.EncryptionField, Value: "AlwaysOn"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BlankScreenField, Value: "true"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.BackupExtensionField, Value: "orig"}},
			},
			expectedBackupFile: targetConfigFilename + ".orig",
		},
		{
			name:             "no backup",
			sourceConfigFile: origTestConfigFilename,
			targetConfigFile: targetConfigFilename,
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: targetConfigFilename}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipReloadField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.EncryptionField, Value: "AlwaysOn"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BlankScreenField, Value: "true"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipBackupField, Value: true}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BackupExtensionField, Value: "orig"}},
			},
		},
		{
			name:             "no changes, so no backup",
			sourceConfigFile: origTestConfigFilename,
			targetConfigFile: targetConfigFilename,
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: targetConfigFilename}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipReloadField, Value: true}},
			},
			// note: this backup file shouldn't exist, but we need the name for the check
			expectedBackupFile: targetConfigFilename + ".bak",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			executor := &realvncserver.RvstExecutor{
				FsManager: &utils.FsManager{},
			}

			configFileContents := makeDummyConfigFile(t, tc.sourceConfigFile, tc.targetConfigFile)
			defer os.Remove(tc.targetConfigFile)

			// convert task values into an actual task
			taskBuilder := rvstbuilder.TaskBuilder{}
			task, err := taskBuilder.Build(
				realvncserver.TaskTypeConfigUpdate,
				"MyPath",
				tc.values,
			)
			require.NoError(t, err)

			rvsTask, ok := task.(*realvncserver.RvsTask)
			require.True(t, ok)

			err = rvsTask.Validate(runtime.GOOS)
			require.NoError(t, err)

			if tc.expectedBackupFile != "" {
				defer os.Remove(tc.expectedBackupFile)
			}

			res := executor.Execute(ctx, rvsTask)
			require.NoError(t, res.Err)
			if rvsTask.Updated {
				backupContents, err := os.ReadFile(tc.expectedBackupFile)
				if rvsTask.SkipBackup {
					assert.ErrorIs(t, err, os.ErrNotExist)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, configFileContents, backupContents)
				}
			} else {
				_, err := os.Stat(tc.expectedBackupFile)
				assert.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}

func makeDummyConfigFile(t *testing.T, sourceConfigFile string, targetConfigFile string) (contents []byte) {
	t.Helper()
	contents, err := os.ReadFile(sourceConfigFile)
	require.NoError(t, err)

	err = os.WriteFile(targetConfigFile, contents, 0600)
	require.NoError(t, err)

	return contents
}
