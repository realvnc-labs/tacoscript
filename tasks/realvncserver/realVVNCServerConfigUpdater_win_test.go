//go:build windows
// +build windows

package realvncserver

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/realvnc-labs/tacoscript/winreg"
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

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	task := &RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "User",
		Encryption: "AlwaysOn",

		mapper:  tracker,
		tracker: tracker,
	}

	err := task.Validate(runtime.GOOS)
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

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")

	setupTask := &RvsTask{
		Path:       "realvnc-server-1",
		ServerMode: "User",
		Encryption: "AlwaysOn",

		mapper:  tracker,
		tracker: tracker,
	}

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	task := &RvsTask{
		Path:       "realvnc-server-2",
		ServerMode: "User",
		Encryption: "PreferOn",

		mapper:  tracker,
		tracker: tracker,
	}

	err = task.Validate(runtime.GOOS)
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

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := &FieldNameStatusTracker{
		NameMap: FieldNameMap{
			"blank_screen": "BlankScreen",
		},
		StatusMap: FieldStatusMap{
			"BlankScreen": FieldStatus{
				HasNewValue: true,
				Clear:       false,
			},
		},
	}

	setupTask := &RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "User",
		BlankScreen: true,

		mapper:  tracker,
		tracker: tracker,
	}

	err := setupTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, setupTask)
	require.NoError(t, res.Err)
	require.True(t, setupTask.Updated)

	tracker = &FieldNameStatusTracker{
		NameMap: FieldNameMap{
			"blank_screen": "BlankScreen",
		},
		StatusMap: FieldStatusMap{
			"BlankScreen": FieldStatus{
				HasNewValue: true,
				Clear:       true,
			},
		},
	}

	clearTask := &RvsTask{
		Path:        "realvnc-server-1",
		ServerMode:  "User",
		BlankScreen: false,

		mapper:  tracker,
		tracker: tracker,
	}

	err = clearTask.Validate(runtime.GOOS)
	require.NoError(t, err)

	res = executor.Execute(ctx, clearTask)
	require.NoError(t, res.Err)
	require.True(t, clearTask.Updated)

	assert.Equal(t, "1 config value change(s) applied", res.Changes["count"])

	found, _, err := winreg.GetValue(HKCUBaseKey, "BlankScreen", winreg.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}
