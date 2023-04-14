package fieldstatus_test

import (
	"fmt"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus/statusbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestTaskWithCombinedNameMapperAndChangeTracker struct {
	Field1 string `taco:"field_1,tracked"`
	Field2 string `taco:"field_2,tracked"`
	Field3 string `taco:"field_3"`

	Require []string `taco:"require"`

	Creates []string `taco:"creates"`
	OnlyIf  []string `taco:"onlyif"`
	Unless  []string `taco:"unless"`

	Shell string `taco:"shell"`

	mapper  fieldstatus.NameMapper
	tracker fieldstatus.Tracker
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetTypeName() string {
	return "test_type"
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetRequirements() []string {
	return t.Require
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetPath() string {
	return "test_path"
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.GetTypeName(), t.GetPath())
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetUnlessCmds() []string {
	return t.Unless
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetCreatesFilesList() []string {
	return t.Creates
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetMapper() (mapper fieldstatus.NameMapper) {
	if t.mapper == nil {
		t.mapper = fieldstatus.NewFieldNameStatusTracker()
	}
	return t.mapper
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) Validate(goos string) error {
	return nil
}

func TestShouldBuildFieldMapForTask(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		mapper:  tracker,
		tracker: tracker,
	}

	statusbuilder.Build(task, tracker, tracker)

	m := task.mapper

	assert.Equal(t, "Field1", m.GetFieldName("field_1"))
	assert.Equal(t, "Field2", m.GetFieldName("field_2"))
	assert.Equal(t, "Field3", m.GetFieldName("field_3"))

	assert.Equal(t, "Creates", m.GetFieldName("creates"))
	assert.Equal(t, "OnlyIf", m.GetFieldName("onlyif"))
	assert.Equal(t, "Unless", m.GetFieldName("unless"))
	assert.Equal(t, "Shell", m.GetFieldName("shell"))

	assert.True(t, tracker.IsTracked("Field1"))
	assert.True(t, tracker.IsTracked("Field2"))
	assert.False(t, tracker.IsTracked("Field3"))
	assert.False(t, tracker.IsTracked("Creates"))
	assert.False(t, tracker.IsTracked("Onlyif"))
}

func TestShouldSetGetFieldStatus(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field1", fieldstatus.FieldStatus{HasNewValue: true, ChangeApplied: true, Clear: true, Tracked: true})
	status, found := task.tracker.GetFieldStatus("Field1")
	assert.True(t, found)
	assert.Equal(t, true, status.HasNewValue)
	assert.Equal(t, true, status.ChangeApplied)
	assert.Equal(t, true, status.Clear)
	assert.Equal(t, true, status.Tracked)
}

func TestShouldFailGetFieldStatus(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	_, found := task.tracker.GetFieldStatus("FieldX")
	assert.False(t, found)
}

func TestShouldHandleHasNewValue(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field1", fieldstatus.FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetHasNewValue("Field1")
	require.NoError(t, err)

	hasChange := task.tracker.HasNewValue("Field1")
	assert.True(t, hasChange)
}

func TestShouldHandleClearChange(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", fieldstatus.FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetClear("Field2")
	require.NoError(t, err)

	fs, found := task.tracker.GetFieldStatus("Field2")
	assert.True(t, found)

	assert.True(t, fs.HasNewValue)
	assert.True(t, fs.Clear)
	assert.False(t, fs.ChangeApplied)
}

func TestShouldSetChangeApplied(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", fieldstatus.FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetChangeApplied("Field2")
	require.NoError(t, err)

	status, found := task.tracker.GetFieldStatus("Field2")
	assert.True(t, found)
	assert.True(t, status.ChangeApplied)
}

func TestShouldSetTracked(t *testing.T) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", fieldstatus.FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false, Tracked: false})

	err := task.tracker.SetTracked("Field2")
	require.NoError(t, err)

	status, found := task.tracker.GetFieldStatus("Field2")
	assert.True(t, found)
	assert.True(t, status.Tracked)
}
