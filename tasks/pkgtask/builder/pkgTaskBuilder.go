package builder

import (
	"fmt"

	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/builder/parser"
	"github.com/realvnc-labs/tacoscript/conv"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
)

type PkgTaskBuilder struct {
}

var pkgTaskParamsFnMap = parser.TaskFieldsParserConfig{
	tasks.NameField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			t := task.(*pkgtask.PkgTask)
			t.Named.Name = fmt.Sprint(val)
			return nil
		},
		FieldName: "Name",
	},
	tasks.NamesField: parser.TaskField{
		ParseFn: func(task tasks.CoreTask, path string, val interface{}) error {
			var names []string
			var err error
			t := task.(*pkgtask.PkgTask)
			names, err = conv.ConvertToValues(val)
			t.Named.Names = names
			return err
		},
		FieldName: "Names",
	},
}

func (fmtb PkgTaskBuilder) Build(typeName, path string, params interface{}) (tasks.CoreTask, error) {
	task := &pkgtask.PkgTask{
		TypeName: typeName,
		Path:     path,
	}

	switch typeName {
	case pkgtask.TaskTypePkgInstalled:
		task.ActionType = pkgtask.ActionInstall
	case pkgtask.TaskTypePkgRemoved:
		task.ActionType = pkgtask.ActionUninstall
	case pkgtask.TaskTypePkgUpgraded:
		task.ActionType = pkgtask.ActionUpdate
	}

	errs := builder.Build(typeName, path, params, task, pkgTaskParamsFnMap)

	return task, errs.ToError()
}
