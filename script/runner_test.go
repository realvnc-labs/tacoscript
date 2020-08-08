package script

import (
	"context"
	"errors"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TaskMock struct {
	ExecResult tasks.ExecutionResult
}

func (tm TaskMock) GetName() string {
	return ""
}

func (tm TaskMock) Execute(ctx context.Context) tasks.ExecutionResult {
	return tm.ExecResult
}

func (tm TaskMock) Validate() error {
	return nil
}

func (tm TaskMock) GetPath() string {
	return ""
}

func TestRunner(t *testing.T) {
	testCases := []struct{
		Scripts tasks.Scripts
		ExpectedOutput []tasks.ExecutionResult
	} {
		{
			ExpectedOutput: []tasks.ExecutionResult{
				{
					StdOut: "some task1",
				},
			},
			Scripts: tasks.Scripts{
				{
					Tasks: []tasks.Task{
						TaskMock{
							ExecResult: tasks.ExecutionResult{
								StdOut: "some task1",
							},
						},
					},
				},
			},
		},
		{
			ExpectedOutput: []tasks.ExecutionResult{
				{
					Err: errors.New("some error"),
				},
				{
					StdOut: "some out",
				},
			},
			Scripts: tasks.Scripts{
				{
					Tasks: []tasks.Task{
						TaskMock{
							ExecResult: tasks.ExecutionResult{
								Err: errors.New("some error"),
							},
						},
					},
				},
				{
					Tasks: []tasks.Task{
						TaskMock{
							ExecResult: tasks.ExecutionResult{
								StdOut: "some out",
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		runr := Runner{}
		actualOutput := runr.Run(context.Background(), testCase.Scripts)
		assert.Equal(t, testCase.ExpectedOutput, actualOutput)
	}
}
