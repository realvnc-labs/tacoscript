//go:build windows
// +build windows

package tasks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudradar-monitoring/tacoscript/utils"
	"github.com/cloudradar-monitoring/tacoscript/winreg"
)

var origHKCUBaseKey string
var testRealVNCBaseKey string

func testSetup(t *testing.T) {
	t.Helper()
	// setup test registry key. assumes User server mode.
	origHKCUBaseKey = HKCUBaseKey
	testRealVNCBaseKey = `HKCU:\Software\RealVNCTest`
	HKCUBaseKey = testRealVNCBaseKey + `\vncserver`
}

func testTeardown(t *testing.T) {
	t.Helper()
	// remove test key and restore base key
	defer func() {
		_ = winreg.DeleteKeyRecursive(testRealVNCBaseKey)
		HKCUBaseKey = origHKCUBaseKey
	}()
}

func TestShouldSetSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RealVNCServerTaskExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	task := &RealVNCServerTask{
		Path:       "realvnc-server-1",
		ServerMode: "User",
		Encryption: "AlwaysOn",
		tracker:    newTrackerWithSingleFieldStatus("encryption", "Encryption"),
	}

	err := task.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := winreg.GetValue(HKCUBaseKey, "Encryption", winreg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "AlwaysOn", regVal)
}

func TestShouldUpdateSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RealVNCServerTaskExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	setupTask := &RealVNCServerTask{
		Path:       "realvnc-server-1",
		ServerMode: "User",
		Encryption: "AlwaysOn",
		tracker:    newTrackerWithSingleFieldStatus("encryption", "Encryption"),
	}

	err := setupTask.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	task := &RealVNCServerTask{
		Path:       "realvnc-server-2",
		ServerMode: "User",
		Encryption: "PreferOn",
		tracker:    newTrackerWithSingleFieldStatus("encryption", "Encryption"),
	}

	err = task.Validate()
	require.NoError(t, err)

	res = executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := winreg.GetValue(HKCUBaseKey, "Encryption", winreg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "PreferOn", regVal)
}

func TestShouldClearSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RealVNCServerTaskExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	setupTask := &RealVNCServerTask{
		Path:        "realvnc-server-1",
		ServerMode:  "User",
		BlankScreen: true,
		tracker: &FieldStatusTracker{
			fieldStatusMap: fieldStatusMap{
				"blank_screen": FieldStatus{
					Name:        "BlankScreen",
					HasNewValue: true,
					Clear:       false,
				},
			},
		},
	}

	err := setupTask.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	clearTask := &RealVNCServerTask{
		Path:        "realvnc-server-1",
		ServerMode:  "User",
		BlankScreen: false,
		tracker: &FieldStatusTracker{
			fieldStatusMap: fieldStatusMap{
				"blank_screen": FieldStatus{
					Name:        "BlankScreen",
					HasNewValue: true,
					Clear:       true,
				},
			},
		},
	}

	err = clearTask.Validate()
	require.NoError(t, err)

	res = executor.Execute(ctx, clearTask)
	require.NoError(t, res.Err)
	require.True(t, clearTask.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, _, err := winreg.GetValue(HKCUBaseKey, "BlankScreen", winreg.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}
