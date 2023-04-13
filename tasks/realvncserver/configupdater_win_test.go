//go:build windows
// +build windows

package realvncserver_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/tasks/support/reg"
	"github.com/realvnc-labs/tacoscript/utils"
)

var origHKLMBaseKey string
var testRealVNCBaseKey string

func testSetup(t *testing.T) {
	// setup test registry key. assumes Service server mode.
	origHKLMBaseKey = realvncserver.HKLMBaseKey
	testRealVNCBaseKey = `HKCU:\Software\RealVNCTest`
	realvncserver.HKLMBaseKey = testRealVNCBaseKey + `\vncserver`
}

func testTeardown(t *testing.T) {
	t.Helper()
	// remove test key and restore base key
	defer func() {
		_ = reg.DeleteKeyRecursive(testRealVNCBaseKey)
		realvncserver.HKLMBaseKey = origHKLMBaseKey
	}()
}

func TestShouldSetSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &realvncserver.RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	task := &realvncserver.RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "Service",
		Encryption: "AlwaysOff",
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := reg.GetValue(realvncserver.HKLMBaseKey, "Encryption", reg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "AlwaysOff", regVal)
}

func TestShouldUpdateSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &realvncserver.RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	setupTask := &realvncserver.RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "Service",
		Encryption: "AlwaysOff",
	}

	setupTask.SetMapper(tracker)
	setupTask.SetTracker(tracker)

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	task := &realvncserver.RvsTask{
		Path:       "realvnc-server-2",
		ServerMode: "Service",
		Encryption: "PreferOn",
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res = executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := reg.GetValue(realvncserver.HKLMBaseKey, "Encryption", reg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "PreferOn", regVal)
}

func TestShouldClearSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &realvncserver.RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		fieldstatus.NameMap{
			"blank_screen": "BlankScreen",
		},
		fieldstatus.StatusMap{
			"BlankScreen": fieldstatus.FieldStatus{
				HasNewValue: true,
				Clear:       false,
			},
		})

	setupTask := &realvncserver.RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "Service",
		BlankScreen: true,
	}

	setupTask.SetMapper(tracker)
	setupTask.SetTracker(tracker)

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	tracker = fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		fieldstatus.NameMap{
			"blank_screen": "BlankScreen",
		},
		fieldstatus.StatusMap{
			"BlankScreen": fieldstatus.FieldStatus{
				HasNewValue: true,
				Clear:       true,
			},
		})

	clearTask := &realvncserver.RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "Service",
		BlankScreen: false,
	}

	clearTask.SetMapper(tracker)
	clearTask.SetTracker(tracker)

	err = clearTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res = executor.Execute(ctx, clearTask)
	require.NoError(t, res.Err)
	require.True(t, clearTask.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, _, err := reg.GetValue(realvncserver.HKLMBaseKey, "BlankScreen", reg.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}
