package realvncserver

import (
	"context"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigReloader struct {
}

func (rl *mockConfigReloader) Reload(rvst *RvsTask) (err error) {
	return nil
}

func TestShouldPerformSimpleConfigParamUpdate(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	ctx := context.Background()

	executor := &RvstExecutor{
		FsManager: &utils.FsManager{},

		Reloader: &mockConfigReloader{},
	}

	tracker := &tasks.FieldNameStatusTracker{
		NameMap: WithNameMap("encryption", "Encryption"),
		StatusMap: tasks.FieldStatusMap{
			"Encryption": tasks.FieldStatus{
				HasNewValue: true,
			},
		},
	}

	task := &RvsTask{
		Path:       "realvnc-server-1",
		Encryption: "AlwaysOn",
		ServerMode: "User",
		Mapper:     tracker,
		Tracker:    tracker,
	}

	if runtime.GOOS != "windows" {
		task.ConfigFile = "../../realvnc/test/realvncserver-config.conf"
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	res := executor.Execute(ctx, task)
	require.NoError(t, res.Err)
	require.True(t, task.Updated)

	assert.Equal(t, res.Comment, "Config updated")
	assert.Equal(t, res.Changes["count"], "1 config value change(s) applied")
}
