package script

import (
	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type positionedRequirement struct {
	previous string
	next     string
}

func SortScriptsRespectingRequirements(scripts tasks.Scripts) {
	requirements, positionsMap := buildRequirementsMap(scripts)

	for {
		allReqMet, req := allRequirementsMet(requirements, positionsMap)
		if allReqMet {
			break
		}

		previousElementIndex := positionsMap[req.previous]
		nexElementIndex := positionsMap[req.next]

		moveScriptToPosition(scripts, previousElementIndex, nexElementIndex)

		positionsMap = make(map[string]int, len(positionsMap))
		for pos, script := range scripts {
			positionsMap[script.ID] = pos
		}
	}
}

func allRequirementsMet(requirements []positionedRequirement, positionsMap map[string]int) (bool, positionedRequirement) {
	for _, req := range requirements {
		if req.previous == req.next {
			continue
		}
		previousPos, ok := positionsMap[req.previous]
		if !ok {
			continue
		}

		nextPos, ok := positionsMap[req.next]
		if !ok {
			continue
		}

		if previousPos > nextPos {
			return false, req
		}
	}

	return true, positionedRequirement{}
}

func buildRequirementsMap(scrpts tasks.Scripts) (req []positionedRequirement, positionsMap map[string]int) {
	req = make([]positionedRequirement, 0)
	positionsMap = make(map[string]int)

	for pos, script := range scrpts {
		positionsMap[script.ID] = pos
		for _, task := range script.Tasks {
			for _, reqName := range task.GetRequirements() {
				req = append(req, positionedRequirement{
					previous: reqName,
					next:     script.ID,
				})
			}
		}
	}

	return req, positionsMap
}

func insertScript(array tasks.Scripts, value tasks.Script, index int) tasks.Scripts {
	return append(array[:index], append(tasks.Scripts{value}, array[index:]...)...)
}

func removeScript(array tasks.Scripts, index int) tasks.Scripts {
	return append(array[:index], array[index+1:]...)
}

func moveScriptToPosition(array tasks.Scripts, srcIndex, dstIndex int) tasks.Scripts {
	if srcIndex > len(array)-1 {
		return tasks.Scripts{}
	}

	value := array[srcIndex]
	return insertScript(removeScript(array, srcIndex), value, dstIndex)
}
