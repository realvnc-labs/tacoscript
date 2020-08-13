package script

import (
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/magiconair/properties/assert"
)

func TestSort(t *testing.T) {
	testCases := []struct {
		scriptsInInput    tasks.Scripts
		expectedScriptIDs []string
	}{
		{
			scriptsInInput: tasks.Scripts{
				{
					ID: "script 7",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 5"},
						},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 1"},
						},
					},
				},
				{
					ID:    "script 5",
					Tasks: []tasks.Task{},
				},
				{
					ID:    "script 6",
					Tasks: []tasks.Task{},
				},
				{
					ID:    "script 4",
					Tasks: []tasks.Task{},
				},
				{
					ID: "script 8",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 1"},
						},
					},
				},
				{
					ID:    "script 1",
					Tasks: []tasks.Task{},
				},
				{
					ID:    "script 9",
					Tasks: []tasks.Task{},
				},
			},
			expectedScriptIDs: []string{"script 5", "script 6", "script 1", "script 7", "script 4", "script 8", "script 9"},
		},
		{
			scriptsInInput: tasks.Scripts{
				{
					ID: "script 1",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 2", "script 3"},
						},
					},
				},
				{
					ID:    "script 2",
					Tasks: []tasks.Task{},
				},
				{
					ID:    "script 3",
					Tasks: []tasks.Task{},
				},
			},
			expectedScriptIDs: []string{"script 2", "script 3", "script 1"},
		},
		{
			scriptsInInput: tasks.Scripts{
				{
					ID: "script 12",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 10"},
						},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 10"},
						},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 11"},
						},
					},
				},
				{
					ID:    "script 10",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 11"},
						},
					},
				},
				{
					ID:    "script 11",
					Tasks: []tasks.Task{},
				},
			},
			expectedScriptIDs: []string{"script 11", "script 10", "script 12"},
		},
	}

	for _, testCase := range testCases {
		actualScripts := testCase.scriptsInInput
		SortScriptsRespectingRequirements(actualScripts)

		actualScriptIDs := make([]string, 0, len(actualScripts))
		for _, script := range actualScripts {
			actualScriptIDs = append(actualScriptIDs, script.ID)
		}

		assert.Equal(t, actualScriptIDs, testCase.expectedScriptIDs)
	}
}
