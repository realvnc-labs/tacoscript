package script

import (
	"context"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
)

type RequirementsTaskMock struct {
	RequirementsToGive []string
	Path               string
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

type cycleErrorExpectation struct {
	messagePrefix  string
	availableParts []string
}

func TestCycleDetection(t *testing.T) {
	testCases := []struct {
		Scripts               tasks.Scripts
		cycleErrorExpectation cycleErrorExpectation
	}{
		{
			Scripts: tasks.Scripts{
				{
					ID: "script one",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script two"},
							Path:               "script one.RequirementsTaskMock[0]",
						},
					},
				},
				{
					ID: "script two",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script one"},
							Path:               "script two.RequirementsTaskMock[0]",
						},
					},
				},
			},
			cycleErrorExpectation: cycleErrorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script one",
					"script two",
				},
			},
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
			cycleErrorExpectation: cycleErrorExpectation{
				messagePrefix: "",
			},
		},
		{
			Scripts: tasks.Scripts{
				{
					ID: "script 6",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 7"},
						},
					},
				},
				{
					ID: "script 7",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 8"},
						},
					},
				},
				{
					ID: "script 8",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
				{
					ID: "script 9",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
			},
			cycleErrorExpectation: cycleErrorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script 6",
					"script 7",
					"script 8",
				},
			},
		},
		{
			Scripts: tasks.Scripts{
				{
					ID: "script 10",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 11"},
						},
					},
				},
				{
					ID: "script 11",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 12", "script 13"},
						},
					},
				},
				{
					ID: "script 13",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{},
						},
					},
				},
			},
			cycleErrorExpectation: cycleErrorExpectation{
				messagePrefix: "",
			},
		},
		{
			Scripts: tasks.Scripts{
				{
					ID: "script 20",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 20"},
							Path:               "script 20 task 1 path 1",
						},
					},
				},
			},
			cycleErrorExpectation: cycleErrorExpectation{
				messagePrefix: "task at path 'script 20 task 1 path 1' cannot require own script 'script 20'",
			},
		},
	}

	for _, testCase := range testCases {
		err := ValidateScripts(testCase.Scripts)
		if testCase.cycleErrorExpectation.messagePrefix == "" {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
			if err == nil {
				continue
			}

			assert.Contains(t, err.Error(), testCase.cycleErrorExpectation.messagePrefix)

			for _, expectedMsgPart := range testCase.cycleErrorExpectation.availableParts {
				assert.Contains(t, err.Error(), expectedMsgPart)
			}
		}
	}
}
