package builder

import (
	builder "github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/winreg"
)

type WinRegTaskBuilder struct {
}

func (wrtb WinRegTaskBuilder) Build(typeName, path string, params interface{}) (tasks.CoreTask, error) {
	task := &winreg.WinRegTask{
		TypeName: typeName,
		Path:     path,
	}

	switch typeName {
	case winreg.TaskTypeWinRegPresent:
		task.ActionType = winreg.ActionWinRegPresent
	case winreg.TaskTypeWinRegAbsent:
		task.ActionType = winreg.ActionWinRegAbsent
	case winreg.TaskTypeWinRegAbsentKey:
		task.ActionType = winreg.ActionWinRegAbsentKey
	}

	errs := builder.Build(typeName, path, params, task, nil)

	return task, errs.ToError()
}
