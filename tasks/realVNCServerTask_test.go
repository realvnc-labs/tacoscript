package tasks

import (
	"context"
	"runtime"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigReloader struct {
}

func (rl *mockConfigReloader) Reload(rvst *RealVNCServerTask) (err error) {
	return nil
}

func TestShouldPerformSimpleConfigParamUpdate(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RealVNCServerTaskExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	task := &RealVNCServerTask{
		Path: "realvnc-server-1",
		tracker: &FieldStatusTracker{
			fieldStatusMap: withNewValue("encryption", "Encryption"),
		},
		Encryption: "AlwaysOn",
		ServerMode: "User",
	}

	if runtime.GOOS != "windows" {
		task.ConfigFile = "../realvnc/test/realvncserver-config.conf"
	}

	err := task.Validate()
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "Config updated")
	assert.Equal(t, res.Changes["count"], "1 config value change(s) applied")
}