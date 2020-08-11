package script

import (
	"context"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
	"testing"
)

type RequirementsTaskMock struct {
	RequirementsToGive []string
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
	return "RequirementsTaskMock"
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
						RequirementsTaskMock{},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script two"},
						},
					},
				},
				{
					ID: "script two",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script one"},
						},
					},
				},
			},
			ExpectedError: "cyclic requirement detected see task at 'script one.RequirementsTaskMock[1]' requires 'script two' at 'script two.RequirementsTaskMock[0]' and vice versa",
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
					ID: "script 4",
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
