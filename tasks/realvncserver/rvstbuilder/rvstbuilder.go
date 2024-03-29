package rvstbuilder

import (
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/shared/builder"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
)

type TaskBuilder struct {
}

func (tb TaskBuilder) Build(typeName, path string, fields interface{}) (t tasks.CoreTask, err error) {
	tracker := fieldstatus.NewFieldNameStatusTracker()
	task := &realvncserver.Task{
		TypeName: typeName,
		Path:     path,
	}

	task.SetMapper(tracker)
	task.SetTracker(tracker)

	errs := builder.Build(typeName, path, fields, task, nil)

	return task, errs.ToError()
}
