package tasks

import (
	"context"
	"fmt"
)

type Executor interface {
	Execute(ctx context.Context, task Task) ExecutionResult
}

type ExecutorRouter struct {
	Executors map[string]Executor
}

func (er ExecutorRouter) GetExecutor(task Task) (Executor, error) {
	e, ok := er.Executors[task.GetName()]
	if !ok {
		return nil, fmt.Errorf("cannot find executor for task %s", task.GetName())
	}

	return e, nil
}
