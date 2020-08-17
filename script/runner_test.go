package script

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
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

func (tm TaskMock) GetRequirements() []string {
	return []string{}
}

func TestRunner(t *testing.T) {
	testCases := []struct {
		Scripts        tasks.Scripts
		ExpectedError string
	}{
		{
			ExpectedError: "",
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
			ExpectedError: "some error",
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
		err := runr.Run(context.Background(), testCase.Scripts)
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
	}
}
