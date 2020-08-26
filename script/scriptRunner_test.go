package script

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
)

type TaskMock struct {
	ID           string
	ExecResult   tasks.ExecutionResult
	Requirements []string
}

func (tm *TaskMock) GetName() string {
	return "TaskMock"
}

func (tm *TaskMock) Validate() error {
	return nil
}

func (tm *TaskMock) GetPath() string {
	return ""
}

func (tm *TaskMock) GetRequirements() []string {
	return tm.Requirements
}

type ExecutorMock struct {
	ExecResult tasks.ExecutionResult
	InputTasks []tasks.Task
}

func (em *ExecutorMock) Execute(ctx context.Context, task tasks.Task) tasks.ExecutionResult {
	em.InputTasks = append(em.InputTasks, task)
	return em.ExecResult
}

func TestRunner(t *testing.T) {
	testCases := []struct {
		Scripts               tasks.Scripts
		ExpectedError         string
		ExecutorMock          *ExecutorMock
		ExpectedExecutedTasks []string
		ExecutorsMapKey       string
	}{
		{
			ExpectedError: "",
			Scripts: tasks.Scripts{
				{
					ID:    "script1",
					Tasks: []tasks.Task{&TaskMock{ID: "123"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: tasks.ExecutionResult{
					StdOut: "some task1",
				},
			},
			ExpectedExecutedTasks: []string{"123"},
			ExecutorsMapKey:       "TaskMock",
		},
		{
			ExpectedError: "some error",
			Scripts: tasks.Scripts{
				{
					ID:    "script3",
					Tasks: []tasks.Task{&TaskMock{ID: "task234"}},
				},
				{
					ID:    "script4",
					Tasks: []tasks.Task{&TaskMock{ID: "task567"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: tasks.ExecutionResult{
					Err: errors.New("some error"),
				},
			},
			ExpectedExecutedTasks: []string{"task234"},
			ExecutorsMapKey:       "TaskMock",
		},
		{
			Scripts: tasks.Scripts{
				{
					ID:    "script5",
					Tasks: []tasks.Task{&TaskMock{ID: "task8"}, &TaskMock{ID: "task9"}},
				},
				{
					ID:    "script6",
					Tasks: []tasks.Task{&TaskMock{ID: "task12"}, &TaskMock{ID: "task13"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: tasks.ExecutionResult{},
			},
			ExpectedExecutedTasks: []string{"task8", "task9", "task12", "task13"},
			ExecutorsMapKey:       "TaskMock",
		},
		{
			ExpectedError: "cannot find executor for task TaskMock",
			Scripts: tasks.Scripts{
				{
					ID:    "script7",
					Tasks: []tasks.Task{&TaskMock{ID: "task10"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: tasks.ExecutionResult{},
			},
			ExpectedExecutedTasks: []string{},
			ExecutorsMapKey:       "someUnknownKey",
		},
		{
			Scripts: tasks.Scripts{
				{
					ID:    "script8",
					Tasks: []tasks.Task{&TaskMock{ID: "task11", Requirements: []string{"script9"}}},
				},
				{
					ID:    "script9",
					Tasks: []tasks.Task{&TaskMock{ID: "task12"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: tasks.ExecutionResult{},
			},
			ExpectedExecutedTasks: []string{"task12", "task11"},
			ExecutorsMapKey:       "TaskMock",
		},
	}

	for _, testCase := range testCases {
		runr := Runner{
			ExecutorRouter: tasks.ExecutorRouter{
				Executors: map[string]tasks.Executor{
					testCase.ExecutorsMapKey: testCase.ExecutorMock,
				},
			},
		}
		err := runr.Run(context.Background(), testCase.Scripts)
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}

		actualExecutedTasks := make([]string, 0, len(testCase.ExecutorMock.InputTasks))
		for _, task := range testCase.ExecutorMock.InputTasks {
			actualExecutedTasks = append(actualExecutedTasks, task.(*TaskMock).ID)
		}

		assert.Equal(t, testCase.ExpectedExecutedTasks, actualExecutedTasks)
	}
}
