package tasks

import (
	"context"
	"time"
)

type Scripts []Script

type Script struct {
	ID    string
	Tasks []Task
}

type Task interface {
	GetName() string
	Execute(ctx context.Context) ExecutionResult
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
}
