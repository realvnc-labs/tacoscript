package statusbuilder

import (
	"reflect"
	"strings"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
)

const (
	TacoStructTag = "taco"
)

func Build(t tasks.CoreTask, mapper fieldstatus.NameMapper, tracker fieldstatus.Tracker) {
	rTaskType := reflect.TypeOf(t)
	rTaskFields := rTaskType.Elem()

	for i := 0; i < rTaskFields.NumField(); i++ {
		fieldName := rTaskFields.Field(i).Name
		tag := rTaskFields.Field(i).Tag
		if tag != "" {
			tagValue := tag.Get(TacoStructTag)
			if tagValue != "" {
				applyTag(mapper, tracker, fieldName, tagValue)
			}
		}
	}
}

func applyTag(mapper fieldstatus.NameMapper, tracker fieldstatus.Tracker, fieldName string, tagValue string) {
	tagValues := strings.Split(tagValue, ",")
	// setup the input key to field name mapping
	inputKey := tagValues[0]
	mapper.SetFieldName(inputKey, fieldName)
	if tracker != nil {
		updateFieldStatus(tracker, fieldName, tagValues)
	}
}

func updateFieldStatus(tracker fieldstatus.Tracker, fieldName string, tagValues []string) {
	// there is a tracker so initialize the field status
	tracker.SetFieldStatus(fieldName, fieldstatus.FieldStatus{})
	if len(tagValues) > 1 {
		// there was a second field in tag, which may mean the field is being tracked
		isTracked := strings.TrimSpace(tagValues[1])
		if strings.EqualFold("true", isTracked) {
			_ = tracker.SetTracked(fieldName)
		}
	}
}
