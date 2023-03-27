package script

import (
	"context"
	"os"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
	"github.com/stretchr/testify/assert"
)

type TaskMock struct {
	ID           string
	ExecResult   executionresult.ExecutionResult
	Requirements []string
	OnlyIf       []string
	Unless       []string
	Creates      []string
}

func (tm *TaskMock) GetTypeName() string {
	return "TaskMock"
}

func (tm *TaskMock) Validate(goos string) error {
	return nil
}

func (tm *TaskMock) GetPath() string {
	return ""
}

func (tm *TaskMock) GetRequirements() []string {
	return tm.Requirements
}

func (tm *TaskMock) GetOnlyIfCmds() []string {
	return tm.OnlyIf
}

func (tm *TaskMock) GetUnlessCmds() []string {
	return tm.Unless
}

func (tm *TaskMock) GetCreatesFilesList() []string {
	return tm.Creates
}

func (tm *TaskMock) IsChangeField(inputKey string) (excluded bool) {
	return false
}

type ExecutorMock struct {
	ExecResult executionresult.ExecutionResult
	InputTasks []tasks.Task
}

func (em *ExecutorMock) Execute(ctx context.Context, task tasks.Task) executionresult.ExecutionResult {
	em.InputTasks = append(em.InputTasks, task)
	return em.ExecResult
}

func TestScriptRunner(t *testing.T) {
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
				tasks.Script{
					ID:    "script1",
					Tasks: []tasks.Task{&TaskMock{ID: "123"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: executionresult.ExecutionResult{
					StdOut: "some task1",
				},
			},
			ExpectedExecutedTasks: []string{"123"},
			ExecutorsMapKey:       "TaskMock",
		},
		{
			Scripts: tasks.Scripts{
				tasks.Script{
					ID:    "script5",
					Tasks: []tasks.Task{&TaskMock{ID: "task8"}, &TaskMock{ID: "task9"}},
				},
				tasks.Script{
					ID:    "script6",
					Tasks: []tasks.Task{&TaskMock{ID: "task12"}, &TaskMock{ID: "task13"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: executionresult.ExecutionResult{},
			},
			ExpectedExecutedTasks: []string{"task8", "task9", "task12", "task13"},
			ExecutorsMapKey:       "TaskMock",
		},
		{
			ExpectedError: "cannot find executor for task TaskMock",
			Scripts: tasks.Scripts{
				tasks.Script{
					ID:    "script7",
					Tasks: []tasks.Task{&TaskMock{ID: "task10"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: executionresult.ExecutionResult{},
			},
			ExpectedExecutedTasks: []string{},
			ExecutorsMapKey:       "someUnknownKey",
		},
		{
			Scripts: tasks.Scripts{
				tasks.Script{
					ID:    "script8",
					Tasks: []tasks.Task{&TaskMock{ID: "task11", Requirements: []string{"script9"}}},
				},
				tasks.Script{
					ID:    "script9",
					Tasks: []tasks.Task{&TaskMock{ID: "task12"}},
				},
			},
			ExecutorMock: &ExecutorMock{
				ExecResult: executionresult.ExecutionResult{},
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
		err := runr.Run(context.Background(), testCase.Scripts, false, os.Stdout)
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
