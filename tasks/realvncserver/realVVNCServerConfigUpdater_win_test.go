//go:build windows
// +build windows

package realvncserver

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/reg"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
)

var origHKLMBaseKey string
var testRealVNCBaseKey string

func testSetup(t *testing.T) {
	// setup test registry key. assumes Service server mode.
	origHKLMBaseKey = HKLMBaseKey
	testRealVNCBaseKey = `HKCU:\Software\RealVNCTest`
	HKLMBaseKey = testRealVNCBaseKey + `\vncserver`
}

func testTeardown(t *testing.T) {
	t.Helper()
	// remove test key and restore base key
	defer func() {
		_ = reg.DeleteKeyRecursive(testRealVNCBaseKey)
		HKLMBaseKey = origHKLMBaseKey
	}()
}

func TestShouldSetSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	task := &RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "Service",
		Encryption: "AlwaysOff",

		Mapper:  tracker,
		Tracker: tracker,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := reg.GetValue(HKLMBaseKey, "Encryption", reg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "AlwaysOff", regVal)
}

func TestShouldUpdateSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	setupTask := &RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "Service",
		Encryption: "AlwaysOff",

		Mapper:  tracker,
		Tracker: tracker,
	}

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	task := &RvsTask{
		Path:       "realvnc-server-2",
		ServerMode: "Service",
		Encryption: "PreferOn",

		Mapper:  tracker,
		Tracker: tracker,
	}

	err = task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res = executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, regVal, err := reg.GetValue(HKLMBaseKey, "Encryption", reg.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "PreferOn", regVal)
}

func TestShouldClearSimpleConfigRegistryParam(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := &tasks.FieldNameStatusTracker{
		NameMap: tasks.FieldNameMap{
			"blank_screen": "BlankScreen",
		},
		StatusMap: tasks.FieldStatusMap{
			"BlankScreen": tasks.FieldStatus{
				HasNewValue: true,
				Clear:       false,
			},
		},
	}

	setupTask := &RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "Service",
		BlankScreen: true,

		Mapper:  tracker,
		Tracker: tracker,
	}

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	tracker = &tasks.FieldNameStatusTracker{
		NameMap: tasks.FieldNameMap{
			"blank_screen": "BlankScreen",
		},
		StatusMap: tasks.FieldStatusMap{
			"BlankScreen": tasks.FieldStatus{
				HasNewValue: true,
				Clear:       true,
			},
		},
	}

	clearTask := &RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "Service",
		BlankScreen: false,

		Mapper:  tracker,
		Tracker: tracker,
	}

	err = clearTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res = executor.Execute(ctx, clearTask)
	require.NoError(t, res.Err)
	require.True(t, clearTask.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, _, err := reg.GetValue(HKLMBaseKey, "BlankScreen", reg.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}
