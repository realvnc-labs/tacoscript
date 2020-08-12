package script

import (
	"context"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
)

type RequirementsTaskMock struct {
	RequirementsToGive []string
	Path string
}

func (rtm RequirementsTaskMock) GetName() string {
	return ""
}

func (rtm RequirementsTaskMock) Execute(ctx context.Context) tasks.ExecutionResult {
	return tasks.ExecutionResult{}
}

func (rtm RequirementsTaskMock) Validate() error {
	return nil
}

func (rtm RequirementsTaskMock) GetPath() string {
	return rtm.Path
}

func (rtm RequirementsTaskMock) GetRequirements() []string {
	return rtm.RequirementsToGive
}

func TestCycleDetection(t *testing.T) {
	testCases := []struct {
		Scripts       tasks.Scripts
		ExpectedError string
	}{
		{
			Scripts: tasks.Scripts{
				{
					ID: "script one",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script two"},
							Path: "script one.RequirementsTaskMock[0]",
						},
					},
				},
				{
					ID: "script two",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script one"},
							Path: "script two.RequirementsTaskMock[0]",
						},
					},
				},
			},
			ExpectedError: "cyclic requirement detected: the task at 'script one.RequirementsTaskMock[0]' " +
				"requires 'script two' which has task at 'script two.RequirementsTaskMock[0]' requiring script 'script one', " +
			"cyclic requirement detected: the task at 'script two.RequirementsTaskMock[0]' " +
				"requires 'script one' which has task at 'script one.RequirementsTaskMock[0]' requiring script 'script two'",
		},
		{
			Scripts: tasks.Scripts{
				{
					ID: "script 3",
					Tasks: []tasks.Task{
						RequirementsTaskMock{},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 4"},
						},
					},
				},
				{
					ID:    "script 4",
					Tasks: []tasks.Task{},
				},
			},
			ExpectedError: "",
		},
	}

	for _, testCase := range testCases {
		err := ValidateScripts(testCase.Scripts)
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
	}
}
