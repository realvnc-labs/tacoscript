package tasks

import (
	"errors"
	"reflect"
)

type fieldNameMap map[string]string

type fieldStatusMap map[string]FieldStatus

type FieldNameStatusTracker struct {
	nameMap   fieldNameMap
	statusMap fieldStatusMap
}

const (
	TacoStructTag = "taco"
)

var (
	ErrFieldNotFoundInTracker = errors.New("field key not found in tracker")
)

func newFieldNameMapper() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap: fieldNameMap{},
	}
}

func newFieldCombinedTracker() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		nameMap:   fieldNameMap{},
		statusMap: fieldStatusMap{},
	}
}

func (tr *FieldNameStatusTracker) BuildFieldMap(t Task) {
	rTaskType := reflect.TypeOf(t)
	rTaskFields := rTaskType.Elem()

	for i := 0; i < rTaskFields.NumField(); i++ {
		fieldName := rTaskFields.Field(i).Name
		tag := rTaskFields.Field(i).Tag
		if tag != "" {
			inputKey := tag.Get(TacoStructTag)
			if inputKey != "" {
				tr.SetFieldName(inputKey, fieldName)
			}
		}
	}
}

func (tr *FieldNameStatusTracker) SetFieldName(fieldKey string, fieldName string) {
	tr.nameMap[fieldKey] = fieldName
}

func (tr *FieldNameStatusTracker) GetFieldName(fk string) (fieldName string) {
	fieldName = tr.nameMap[fk]
	return fieldName
}

func (tr *FieldNameStatusTracker) Init() {
	tr.statusMap = fieldStatusMap{}
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
		HasNewValue:   true,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         true,
	}
	return nil
}

func (tr *FieldNameStatusTracker) SetChangeApplied(fieldName string) (err error) {
	existingStatus, found := tr.GetFieldStatus(fieldName)
	if !found {
		return ErrFieldNotFoundInTracker
	}
	tr.statusMap[fieldName] = FieldStatus{
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
