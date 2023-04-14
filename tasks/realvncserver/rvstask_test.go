package realvncserver_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigReloader struct {
}

func (rl *mockConfigReloader) Reload(rvst *realvncserver.Task) (err error) {
	return nil
}

func TestShouldPerformSimpleConfigParamUpdate(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &realvncserver.Executor{
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

	task := &realvncserver.Task{
		Path:       "realvnc-server-1",
		Encryption: "AlwaysOn",
		ServerMode: "Service",
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	if runtime.GOOS != "windows" {
		task.ConfigFile = "../../testdata/realvncserver-config.conf"
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "Config updated")
	assert.Equal(t, res.Changes["count"], "1 config value change(s) applied")
}
