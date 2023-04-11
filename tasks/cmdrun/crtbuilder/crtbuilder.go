package crtbuilder

import (
	"fmt"

	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/builder/parser"
	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
)

type TaskBuilder struct {
}

var cmdRunTaskParamsFnMap = parser.TaskFieldsParserConfig{
	tasks.NameField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			t := task.(*cmdrun.CrTask)
			t.Named.Name = fmt.Sprint(val)
			return nil
		},
		FieldName: "Name",
	},
	tasks.NamesField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var err error
			t := task.(*cmdrun.CrTask)
			t.Named.Names, err = conv.ConvertToValues(val)
			return err
		},
		FieldName: "Names",
	},
	tasks.EnvField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var err error
			t := task.(*cmdrun.CrTask)
			t.Envs, err = conv.ConvertToKeyValues(val, path)
			return err
		},
		FieldName: "Env",
	},
}

func (tb TaskBuilder) Build(typeName, path string, params interface{}) (t tasks.CoreTask, err error) {
	task := &cmdrun.CrTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := builder.Build(typeName, path, params, task, cmdRunTaskParamsFnMap)

	return task, errs.ToError()
}
