package tasks

import (
	"errors"
	"reflect"
	"strings"
)

type FieldNameMapper interface {
	BuildFieldMap(t CoreTask)
	GetFieldName(fk string) (fieldName string)
	SetFieldName(fk string, fieldName string)
}

type FieldStatus struct {
	Tracked       bool // used to indicate the field is being tracked
	HasNewValue   bool // means that a field value has been included in the task and should applied
	ChangeApplied bool // indicates that a new value has been applied to a target
	Clear         bool // can be set to remove a value from a target
}

type FieldStatusTracker interface {
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

type FieldNameMap map[string]string

type FieldStatusMap map[string]FieldStatus

type FieldNameStatusTracker struct {
	NameMap   FieldNameMap
	StatusMap FieldStatusMap
}

const (
	TacoStructTag = "taco"
)

var (
	ErrFieldNotFoundInTracker = errors.New("field key not found in tracker")
)

func NewFieldNameMapper() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		NameMap: FieldNameMap{},
	}
}

func NewFieldCombinedTracker() (tr *FieldNameStatusTracker) {
	return &FieldNameStatusTracker{
		NameMap:   FieldNameMap{},
		StatusMap: FieldStatusMap{},
	}
}

func (tr *FieldNameStatusTracker) HasStatusTracker() (has bool) {
	return tr.StatusMap != nil
}

func (tr *FieldNameStatusTracker) BuildFieldMap(t CoreTask) {
	rTaskType := reflect.TypeOf(t)
	rTaskFields := rTaskType.Elem()

	for i := 0; i < rTaskFields.NumField(); i++ {
		fieldName := rTaskFields.Field(i).Name
		tag := rTaskFields.Field(i).Tag
		if tag != "" {
			tagValue := tag.Get(TacoStructTag)
			if tagValue != "" {
				tagValues := strings.Split(tagValue, ",")
				// setup the input key to field name mapping
				inputKey := tagValues[0]
				tr.SetFieldName(inputKey, fieldName)
				if tr.HasStatusTracker() {
					// there is a tracker so initialize the field status
					tr.SetFieldStatus(fieldName, FieldStatus{})
					if len(tagValues) > 1 {
						// there was a second field in tag, which may mean the field is being tracked
						isTracked := strings.TrimSpace(tagValues[1])
						if strings.EqualFold("true", isTracked) {
							_ = tr.SetTracked(fieldName)
						}
					}
				}
			}
		}
	}
}

func (tr *FieldNameStatusTracker) SetFieldName(fieldKey string, fieldName string) {
	tr.NameMap[fieldKey] = fieldName
}

func (tr *FieldNameStatusTracker) GetFieldName(fk string) (fieldName string) {
	fieldName = tr.NameMap[fk]
	return fieldName
}

func (tr *FieldNameStatusTracker) Init() {
	tr.StatusMap = FieldStatusMap{}
}

func (tr *FieldNameStatusTracker) GetFieldStatus(fieldName string) (status FieldStatus, found bool) {
	status, found = tr.StatusMap[fieldName]
	return status, found
}

func (tr *FieldNameStatusTracker) SetFieldStatus(fieldName string, status FieldStatus) {
	tr.StatusMap[fieldName] = status
}

func (tr *FieldNameStatusTracker) HasNewValue(fieldName string) (hasNew bool) {
	fieldStatus, found := tr.StatusMap[fieldName]
	if found {
		return fieldStatus.HasNewValue
	}
	return false
}

func (tr *FieldNameStatusTracker) ShouldClear(fieldName string) (should bool) {
	fieldStatus, found := tr.StatusMap[fieldName]
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
	tr.StatusMap[fieldName] = FieldStatus{
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
	tr.StatusMap[fieldName] = FieldStatus{
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
	tr.StatusMap[fieldName] = FieldStatus{
		Tracked:       true,
		HasNewValue:   existingStatus.HasNewValue,
		ChangeApplied: existingStatus.ChangeApplied,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldNameStatusTracker) IsTracked(fieldName string) (flagged bool) {
	fieldStatus, found := tr.StatusMap[fieldName]
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
	tr.StatusMap[fieldName] = FieldStatus{
		Tracked:       existingStatus.Tracked,
		HasNewValue:   existingStatus.HasNewValue,
		ChangeApplied: true,
		Clear:         existingStatus.Clear,
	}
	return nil
}

func (tr *FieldNameStatusTracker) WithNewValues(applyFn func(fn string, fs FieldStatus) (err error)) (err error) {
	for fn, fs := range tr.StatusMap {
		if fs.HasNewValue {
			err = applyFn(fn, fs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
