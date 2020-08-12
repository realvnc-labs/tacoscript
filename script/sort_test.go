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
					ID: "script 4",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 5"},
						},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
				{
					ID:    "script 5",
					Tasks: []tasks.Task{},
				},
				{
					ID: "script 6",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 7"},
						},
					},
				},
				{
					ID:    "script 7",
					Tasks: []tasks.Task{},
				},
			},
			expectedScriptIDs: []string{"script 5", "script 7", "script 6", "script 4"},
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
