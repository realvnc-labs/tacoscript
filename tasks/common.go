package tasks

import (
	"fmt"
	"time"
)

type Scripts []Script

type Script struct {
	ID    string
	Tasks []Task
}

type Task interface {
	GetName() string
	Validate() error
	GetPath() string
	GetRequirements() []string
}

type ExecutionResult struct {
	Err       error
	Duration  time.Duration
	StdErr    string
	StdOut    string
	IsSkipped bool
	Pids      []int
}

func (tr *ExecutionResult) String() string {
	if tr.Err != nil {
		return fmt.Sprintf(`Execution failed: %v, StdErr: %s, Took: %v, StdOut: %s`, tr.Err, tr.StdErr, tr.Duration, tr.StdOut)
	}

	if tr.IsSkipped {
		return fmt.Sprintf(`Execution is Skipped, StdOut: %s, StdErr: %s, Took: %v`, tr.StdOut, tr.StdErr, tr.Duration)
	}

	return fmt.Sprintf(`Execution success, StdOut: %s, StdErr: %s, Took: %s`, tr.StdOut, tr.StdErr, tr.Duration)
}

// returns true if task succeeded or was skipped
func (tr *ExecutionResult) Succeeded() bool {
	return tr.Err == nil
}
