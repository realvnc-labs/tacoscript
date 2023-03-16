package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBuildFieldMapForTask(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	m := task.tracker.fieldStatusMap

	assert.Equal(t, "Encryption", m["encryption"].Name)
	assert.Equal(t, "Authentication", m["authentication"].Name)
	assert.Equal(t, "Permissions", m["permissions"].Name)
	assert.Equal(t, "QueryConnect", m["query_connect"].Name)
	assert.Equal(t, "QueryOnlyIfLoggedOn", m["query_only_if_logged_on"].Name)
	assert.Equal(t, "QueryConnectTimeoutSecs", m["query_connect_timeout"].Name)
	assert.Equal(t, "BlankScreen", m["blank_screen"].Name)
	assert.Equal(t, "ConnNotifyTimeoutSecs", m["conn_notify_timeout"].Name)
	assert.Equal(t, "ConnNotifyAlways", m["conn_notify_always"].Name)
	assert.Equal(t, "IdleTimeoutSecs", m["idle_timeout"].Name)
	assert.Equal(t, "Log", m["log"].Name)
	assert.Equal(t, "CaptureMethod", m["capture_method"].Name)
	assert.Equal(t, "ConfigFile", m["config_file"].Name)
	assert.Equal(t, "ServerMode", m["server_mode"].Name)
	assert.Equal(t, "Require", m["require"].Name)
	assert.Equal(t, "Creates", m["creates"].Name)
	assert.Equal(t, "OnlyIf", m["onlyif"].Name)
	assert.Equal(t, "Unless", m["unless"].Name)
	assert.Equal(t, "Shell", m["shell"].Name)
}

func TestShouldSetGetFieldStatus(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)

	task.tracker.SetFieldStatus("log", FieldStatus{Name: "Log", HasNewValue: true, ChangeApplied: true, Clear: true})
	status, found := task.tracker.GetFieldStatus("log")
	assert.True(t, found)
	assert.Equal(t, "Log", status.Name)
	assert.Equal(t, true, status.HasNewValue)
	assert.Equal(t, true, status.ChangeApplied)
	assert.Equal(t, true, status.Clear)
}

func TestShouldGetFieldStatusByName(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	fs, fk, found := task.tracker.GetFieldStatusByName("Encryption")
	assert.True(t, found)
	assert.Equal(t, fk, "encryption")
	assert.Equal(t, "Encryption", fs.Name)
}

func TestShouldFailGetFieldStatusByName(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	_, _, found := task.tracker.GetFieldStatusByName("Encryption1")
	assert.False(t, found)
}

func TestShouldHandleHasNewValue(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	task.tracker.SetFieldStatus("encryption", FieldStatus{Name: "Encryption", HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetHasNewValue("encryption")
	require.NoError(t, err)

	hasChange := task.tracker.HasNewValue("encryption")
	assert.True(t, hasChange)
}

func TestShouldHandleClearChange(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	task.tracker.SetFieldStatus("encryption", FieldStatus{Name: "Encryption", HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetClear("encryption")
	require.NoError(t, err)

	fs, found := task.tracker.GetFieldStatus("encryption")
	assert.True(t, found)

	assert.True(t, fs.HasNewValue)
	assert.True(t, fs.Clear)
	assert.False(t, fs.ChangeApplied)
}

func TestShouldSetChangeApplied(t *testing.T) {
	task := &RealVNCServerTask{
		tracker: newFieldStatusTracker(),
	}

	task.tracker.BuildFieldMap(task)
	task.tracker.SetFieldStatus("encryption", FieldStatus{Name: "Encryption", HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetChangeApplied("encryption")
	require.NoError(t, err)

	status, found := task.tracker.GetFieldStatus("encryption")
	assert.True(t, found)
	assert.True(t, status.ChangeApplied)
}
