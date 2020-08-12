package script

import (
	"sort"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

func SortScriptsRespectingRequirements(scrptsS tasks.Scripts) {
	requirementsMap := buildRequirementsMap(scrptsS)

	sort.Slice(scrptsS, func(leftIndex, rightIndex int) bool {
		if leftIndex > len(scrptsS)-1 || rightIndex > len(scrptsS)-1 {
			return false
		}

		leftScript := scrptsS[leftIndex]
		rightScript := scrptsS[rightIndex]

		// means left script is required by the right script
		if _, ok := requirementsMap[leftScript.ID][rightScript.ID]; ok {
			return true // since left is required by right it should go first
		}

		return false
	})
}

func buildRequirementsMap(scrpts tasks.Scripts) map[string]map[string]bool {
	reqMap := make(map[string]map[string]bool)

	for _, script := range scrpts {
		for _, task := range script.Tasks {
			for _, reqName := range task.GetRequirements() {
				if _, ok := reqMap[reqName]; !ok {
					reqMap[reqName] = make(map[string]bool)
				}

				reqMap[reqName][script.ID] = true
			}
		}
	}

	return reqMap
}
