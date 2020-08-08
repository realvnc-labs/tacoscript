package script

import (
	"context"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type Runner struct {

}

func (r Runner) Run(ctx context.Context, scripts tasks.Scripts) []tasks.ExecutionResult {
	results := make([]tasks.ExecutionResult, 0, len(scripts))
	for _, script := range scripts {
		for _, task := range script.Tasks {
			results = append(results, task.Execute(ctx))
		}
	}

	return results
}
