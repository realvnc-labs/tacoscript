package tasks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestTaskWithCombinedNameMapperAndChangeTracker struct {
	Field1 string `taco:"field_1"`
	Field2 string `taco:"field_2"`
	Field3 string `taco:"field_3"`

	Require []string `taco:"require"`

	Creates []string `taco:"creates"`
	OnlyIf  []string `taco:"onlyif"`
	Unless  []string `taco:"unless"`

	Shell string `taco:"shell"`

	mapper  FieldNameMapper
	tracker FieldStatusTracker
}

var (
	TestNoChangeFields = []string{"Field3"}
)

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetTracker() (tracker FieldStatusTracker) {
	return t.tracker
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) IsChangeField(fieldName string) (excluded bool) {
	for _, noChangeField := range TestNoChangeFields {
		if fieldName == noChangeField {
			return false
		}
	}
	return true
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

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) GetMapper() (mapper FieldNameMapper) {
	if t.mapper == nil {
		t.mapper = NewFieldCombinedTracker()
	}
	return t.mapper
}

func (t *TestTaskWithCombinedNameMapperAndChangeTracker) Validate(goos string) error {
	return nil
}

func TestShouldBuildFieldMapForTask(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		mapper:  tracker,
		tracker: tracker,
	}

	task.mapper.BuildFieldMap(task)
	m := task.mapper

	assert.Equal(t, "Field1", m.GetFieldName("field_1"))
	assert.Equal(t, "Field2", m.GetFieldName("field_2"))
	assert.Equal(t, "Field3", m.GetFieldName("field_3"))

	assert.Equal(t, "Creates", m.GetFieldName("creates"))
	assert.Equal(t, "OnlyIf", m.GetFieldName("onlyif"))
	assert.Equal(t, "Unless", m.GetFieldName("unless"))
	assert.Equal(t, "Shell", m.GetFieldName("shell"))
}

func TestShouldSetGetFieldStatus(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field1", FieldStatus{HasNewValue: true, ChangeApplied: true, Clear: true})
	status, found := task.tracker.GetFieldStatus("Field1")
	assert.True(t, found)
	assert.Equal(t, true, status.HasNewValue)
	assert.Equal(t, true, status.ChangeApplied)
	assert.Equal(t, true, status.Clear)
}

func TestShouldFailGetFieldStatus(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	_, found := task.tracker.GetFieldStatus("FieldX")
	assert.False(t, found)
}

func TestShouldHandleHasNewValue(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field1", FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetHasNewValue("Field1")
	require.NoError(t, err)

	hasChange := task.tracker.HasNewValue("Field1")
	assert.True(t, hasChange)
}

func TestShouldHandleClearChange(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetClear("Field2")
	require.NoError(t, err)

	fs, found := task.tracker.GetFieldStatus("Field2")
	assert.True(t, found)

	assert.True(t, fs.HasNewValue)
	assert.True(t, fs.Clear)
	assert.False(t, fs.ChangeApplied)
}

func TestShouldSetChangeApplied(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	err := task.tracker.SetChangeApplied("Field2")
	require.NoError(t, err)

	status, found := task.tracker.GetFieldStatus("Field2")
	assert.True(t, found)
	assert.True(t, status.ChangeApplied)
}

func TestShouldKnowIfFieldIsNotChangeField(t *testing.T) {
	tracker := NewFieldCombinedTracker()
	task := &TestTaskWithCombinedNameMapperAndChangeTracker{
		tracker: tracker,
	}

	task.tracker.SetFieldStatus("Field2", FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})
	task.tracker.SetFieldStatus("Field3", FieldStatus{HasNewValue: false, ChangeApplied: false, Clear: false})

	isChangeField := task.IsChangeField("Field2")
	assert.True(t, isChangeField)
	isChangeField = task.IsChangeField("Field3")
	assert.False(t, isChangeField)
}
