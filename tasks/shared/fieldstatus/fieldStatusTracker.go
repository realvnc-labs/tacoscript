package fieldstatus

import (
	"errors"
)

type NameMapper interface {
	GetFieldName(fk string) (fieldName string)
	SetFieldName(fk string, fieldName string)
}

type FieldStatus struct {
	Tracked       bool // used to indicate the field is being tracked
	HasNewValue   bool // means that a field value has been included in the task and should applied
	ChangeApplied bool // indicates that a new value has been applied to a target
	Clear         bool // can be set to remove a value from a target
}

type Tracker interface {
	GetFieldStatus(fieldName string) (status FieldStatus, found bool)
	SetFieldStatus(fieldName string, status FieldStatus)
	SetTracked(fieldName string) (err error)
	IsTracked(fieldName string) (flagged bool)
	SetHasNewValue(fieldName string) (err error)
	HasNewValue(fieldName string) (hasNew bool)
	SetClear(fieldName string) (err error)
	ShouldClear(fieldName string) (should bool)
	SetChangeApplied(fieldName string) (err error)
	WithNewValues(applyFn func(fieldName string, fs FieldStatus) (err error)) (err error)
}

type NameMap map[string]string

type StatusMap map[string]FieldStatus

type FieldNameStatusTracker struct {
	nameMap   NameMap
	statusMap StatusMap
}

var (
	ErrFieldNotFoundInTracker = errors.New("field key not found in tracker")
)

func NewFieldNameMapper() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap: NameMap{},
	}
}

func NewFieldNameStatusTracker() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap:   NameMap{},
		statusMap: StatusMap{},
	}
}

func NewFieldNameMapperWithMap(nameMap NameMap) (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap: nameMap,
	}
}

func NewFieldNameStatusTrackerWithMapAndStatus(nameMap NameMap, statusMap StatusMap) (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap:   nameMap,
		statusMap: statusMap,
	}
}

func (tr *FieldNameStatusTracker) HasStatusTracker() (has bool) {
	return tr.statusMap != nil
}

func (tr *FieldNameStatusTracker) SetFieldName(fieldKey string, fieldName string) {
	tr.nameMap[fieldKey] = fieldName
}

func (tr *FieldNameStatusTracker) GetFieldName(fk string) (fieldName string) {
	fieldName = tr.nameMap[fk]
	return fieldName
}

func (tr *FieldNameStatusTracker) Init() {
	tr.statusMap = StatusMap{}
}

func (tr *FieldNameStatusTracker) GetFieldStatus(fieldName string) (status FieldStatus, found bool) {
	status, found = tr.statusMap[fieldName]
	return status, found
}

func (tr *FieldNameStatusTracker) SetFieldStatus(fieldName string, status FieldStatus) {
	tr.statusMap[fieldName] = status
}

func (tr *FieldNameStatusTracker) HasNewValue(fieldName string) (hasNew bool) {
	fieldStatus, found := tr.statusMap[fieldName]
	if found {
		return fieldStatus.HasNewValue
	}
	return false
}

func (tr *FieldNameStatusTracker) ShouldClear(fieldName string) (should bool) {
	fieldStatus, found := tr.statusMap[fieldName]
	if found {
		return fieldStatus.Clear
	}
	return false
}

func (tr *FieldNameStatusTracker) SetHasNewValue(fieldName string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldName)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.statusMap[fieldName] = FieldStatus{
		Tracked:       existingStatus.Tracked,
		HasNewValue:   true,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldNameStatusTracker) SetClear(fieldName string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldName)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.statusMap[fieldName] = FieldStatus{
		Tracked:       existingStatus.Tracked,
		HasNewValue:   true,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         true,
	}
	return nil
}

func (tr *FieldNameStatusTracker) SetTracked(fieldName string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldName)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.statusMap[fieldName] = FieldStatus{
		Tracked:       true,
		HasNewValue:   existingStatus.HasNewValue,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldNameStatusTracker) IsTracked(fieldName string) (flagged bool) {
	fieldStatus, found := tr.statusMap[fieldName]
	if found {
		return fieldStatus.Tracked
	}
	return false
}

func (tr *FieldNameStatusTracker) SetChangeApplied(fieldName string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldName)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.statusMap[fieldName] = FieldStatus{
		Tracked:       existingStatus.Tracked,
		HasNewValue:   existingStatus.HasNewValue,
		ChangeApplied: true,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldNameStatusTracker) WithNewValues(applyFn func(fn string, fs FieldStatus) (err error)) (err error) {
	for fn, fs := range tr.statusMap {
		if fs.HasNewValue {
			err = applyFn(fn, fs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
