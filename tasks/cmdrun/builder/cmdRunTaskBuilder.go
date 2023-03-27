package builder

import (
	"fmt"

	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/builder/parser"
	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
)

type CmdRunTaskBuilder struct {
}

var cmdRunTaskParamsFnMap = parser.TaskFieldsParserConfig{
	tasks.NameField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			t := task.(*cmdrun.CmdRunTask)
			t.Named.Name = fmt.Sprint(val)
			return nil
		},
		FieldName: "Name",
	},
	tasks.NamesField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var err error
			t := task.(*cmdrun.CmdRunTask)
			t.Named.Names, err = conv.ConvertToValues(val)
			return err
		},
		FieldName: "Names",
	},
	tasks.EnvField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var err error
			t := task.(*cmdrun.CmdRunTask)
			t.Envs, err = conv.ConvertToKeyValues(val, path)
			return err
		},
		FieldName: "Env",
	},
}

func (crtb CmdRunTaskBuilder) Build(typeName, path string, params interface{}) (t tasks.CoreTask, err error) {
	task := &cmdrun.CmdRunTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := builder.Build(typeName, path, params, task, cmdRunTaskParamsFnMap)

	return task, errs.ToError()
}
