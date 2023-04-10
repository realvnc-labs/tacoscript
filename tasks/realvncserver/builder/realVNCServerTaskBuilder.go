package builder

import (
	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/fieldstatus"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
)

type RealVNCServerTaskBuilder struct {
}

func (tb RealVNCServerTaskBuilder) Build(typeName, path string, fields interface{}) (t tasks.CoreTask, err error) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &realvncserver.RvsTask{
		TypeName: typeName,
		Path:     path,
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	errs := builder.Build(typeName, path, fields, task, nil)

	return task, errs.ToError()
}
