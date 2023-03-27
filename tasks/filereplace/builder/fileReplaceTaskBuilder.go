package builder

import (
	builder "github.com/realvnc-labs/tacoscript/builder"
	tasks "github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filereplace"
)

type FileReplaceTaskBuilder struct {
}

func (frtb FileReplaceTaskBuilder) Build(typeName, path string, params interface{}) (t tasks.CoreTask, err error) {
	task := &filereplace.FileReplaceTask{
		TypeName: typeName,
		Path:     path,
	}

	errs := builder.Build(typeName, path, params, task, nil)

	return task, errs.ToError()
}
