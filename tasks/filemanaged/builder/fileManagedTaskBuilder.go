package builder

import (
	"database/sql"
	"fmt"

	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/builder/parser"
	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	"github.com/realvnc-labs/tacoscript/utils"
)

type FileManagedTaskBuilder struct {
}

var FileManagedTaskParamsFnMap = parser.TaskFieldsParserConfig{
	tasks.ModeField: parser.TaskField{
		ParseFn: func(task tasks.Task, path string, val interface{}) error {
			var err error
			t := task.(*filemanaged.FileManagedTask)
			t.Mode, err = conv.ConvertToFileMode(val)
			return err
		},
		FieldName: "Mode",
	},
	tasks.SourceField: parser.TaskField{
		ParseFn: func(task tasks.Task, path string, val interface{}) error {
			t := task.(*filemanaged.FileManagedTask)
			t.Source = utils.ParseLocation(fmt.Sprint(val))
			return nil
		},
		FieldName: "Source",
	},
	tasks.ContentsField: parser.TaskField{
		ParseFn: func(task tasks.Task, path string, val interface{}) error {
			t := task.(*filemanaged.FileManagedTask)
			t.Contents = parseContentsField(val)
			return nil
		},
		FieldName: "Contents",
	},
}

func (fmtb FileManagedTaskBuilder) Build(typeName, path string, params interface{}) (tasks.Task, error) {
	task := &filemanaged.FileManagedTask{
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
