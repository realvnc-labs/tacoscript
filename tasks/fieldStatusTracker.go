package tasks

import (
	"errors"
	"reflect"
)

type FieldStatus struct {
	Name          string
	HasNewValue   bool
	ChangeApplied bool
	Clear         bool
}

type fieldStatusMap map[string]FieldStatus

type FieldStatusTracker struct {
	fieldStatusMap fieldStatusMap
}

const (
	TacoStructTag = "taco"
)

var (
	ErrFieldNotFoundInTracker = errors.New("field key not found in tracker")
)

func newFieldStatusTracker() (tr *FieldStatusTracker) {
	return &FieldStatusTracker{
		fieldStatusMap: map[string]FieldStatus{},
	}
}

func (tr *FieldStatusTracker) BuildFieldMap(t Task) {
	rTaskType := reflect.TypeOf(t)
	rTaskFields := rTaskType.Elem()

	for i := 0; i < rTaskFields.NumField(); i++ {
		fieldName := rTaskFields.Field(i).Name
		tag := rTaskFields.Field(i).Tag
		if tag != "" {
			inputKey := tag.Get(TacoStructTag)
			if inputKey != "" {
				tr.SetFieldStatus(inputKey, FieldStatus{Name: fieldName, HasNewValue: false, Clear: false})
			}
		}
	}
}

func (tr *FieldStatusTracker) SetFieldStatus(fieldKey string, status FieldStatus) {
	tr.fieldStatusMap[fieldKey] = status
}

func (tr *FieldStatusTracker) GetFieldStatus(fieldKey string) (status FieldStatus, found bool) {
	status, found = tr.fieldStatusMap[fieldKey]
	return status, found
}

func (tr *FieldStatusTracker) GetFieldStatusByName(fieldName string) (status FieldStatus, fieldKey string, found bool) {
	for fk, fs := range tr.fieldStatusMap {
		if fs.Name == fieldName {
			return fs, fk, true
		}
	}
	return status, "", false
}

func (tr *FieldStatusTracker) HasNewValue(fieldKey string) (hasNew bool) {
	fieldStatus, found := tr.fieldStatusMap[fieldKey]
	if found {
		return fieldStatus.HasNewValue
	}
	return false
}

func (tr *FieldStatusTracker) ShouldClear(fieldKey string) (should bool) {
	fieldStatus, found := tr.fieldStatusMap[fieldKey]
	if found {
		return fieldStatus.Clear
	}
	return false
}

func (tr *FieldStatusTracker) SetHasNewValue(fieldKey string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldKey)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.fieldStatusMap[fieldKey] = FieldStatus{
		Name:          existingStatus.Name,
		HasNewValue:   true,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldStatusTracker) SetClear(fieldKey string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldKey)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.fieldStatusMap[fieldKey] = FieldStatus{
		Name:          existingStatus.Name,
		HasNewValue:   true,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         true,
	}
	return nil
}

func (tr *FieldStatusTracker) SetChangeApplied(fieldKey string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldKey)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.fieldStatusMap[fieldKey] = FieldStatus{
		Name:          existingStatus.Name,
		HasNewValue:   existingStatus.HasNewValue,
		ChangeApplied: true,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldStatusTracker) WithNewValues(applyFn func(fk string, fs FieldStatus) (err error)) (err error) {
	for fk, fs := range tr.fieldStatusMap {
		if fs.HasNewValue {
			err = applyFn(fk, fs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
