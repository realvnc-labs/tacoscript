package builder

import (
	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
)

type RealVNCServerTaskBuilder struct {
}

func (tb RealVNCServerTaskBuilder) Build(typeName, path string, fields interface{}) (t tasks.CoreTask, err error) {
	tracker := tasks.NewFieldCombinedTracker()
	task := &realvncserver.RvsTask{
		TypeName: typeName,
		Path:     path,
		Mapper:   tracker,
		Tracker:  tracker,
	}

	errs := builder.Build(typeName, path, fields, task, nil)

	return task, errs.ToError()
}
