package fmtbuilder

import (
	"database/sql"
	"fmt"

	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	"github.com/realvnc-labs/tacoscript/tasks/shared/builder"
	"github.com/realvnc-labs/tacoscript/tasks/shared/builder/parser"
	"github.com/realvnc-labs/tacoscript/utils"
)

type TaskBuilder struct {
}

var FileManagedTaskParamsFnMap = parser.TaskFieldsParserConfig{
	tasks.ModeField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var err error
			t := task.(*filemanaged.Task)
			t.Mode, err = conv.ConvertToFileMode(val)
			return err
		},
		FieldName: "Mode",
	},
	tasks.SourceField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			t := task.(*filemanaged.Task)
			t.Source = utils.ParseLocation(fmt.Sprint(val))
			return nil
		},
		FieldName: "Source",
	},
	tasks.ContentsField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			t := task.(*filemanaged.Task)
			t.Contents = parseContentsField(val)
			return nil
		},
		FieldName: "Contents",
	},
}

func (tb TaskBuilder) Build(typeName, path string, params interface{}) (tasks.CoreTask, error) {
	task := &filemanaged.Task{
		TypeName: typeName,
		Path:     path,
		Replace:  true,
	}

	errs := builder.Build(typeName, path, params, task, FileManagedTaskParamsFnMap)

	return task, errs.ToError()
}

func parseContentsField(val interface{}) sql.NullString {
	isValid := false
	if val != nil {
		isValid = true
	}
	return sql.NullString{
		String: fmt.Sprint(val),
		Valid:  isValid,
	}
}
