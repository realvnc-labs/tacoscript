package tasks

import (
	"context"
	"fmt"

	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
)

type Executor interface {
	Execute(ctx context.Context, task Task) executionresult.ExecutionResult
}

type ExecutorRouter struct {
	Executors map[string]Executor
}

func (er ExecutorRouter) GetExecutor(task Task) (Executor, error) {
	e, ok := er.Executors[task.GetTypeName()]
	if !ok {
		return nil, fmt.Errorf("cannot find executor for task %s", task.GetTypeName())
	}

	return e, nil
}
