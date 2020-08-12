package script

import (
	"fmt"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/cloudradar-monitoring/tacoscript/utils"
	"github.com/elliotchance/orderedmap"
)

func isCyclic(curScriptID string,
	scriptIDToRequiredScriptIDsMap map[string][]string,
	visited map[string]bool,
	requestStack,
	cyclicItems *orderedmap.OrderedMap) bool {
	if isInRequestStack, ok := requestStack.Get(curScriptID); ok && isInRequestStack.(bool) {
		cyclicItems.Set(curScriptID, true)
		return true
	}

	if isVisited, ok := visited[curScriptID]; ok && isVisited {
		return false
	}

	visited[curScriptID] = true
	requestStack.Set(curScriptID, true)
	cyclicItems.Set(curScriptID, true)

	if requirements, ok := scriptIDToRequiredScriptIDsMap[curScriptID]; ok {
		for _, requirement := range requirements {
			isCyclic := isCyclic(requirement, scriptIDToRequiredScriptIDsMap, visited, requestStack, cyclicItems)
			if isCyclic {
				return true
			}
		}
	}

	requestStack.Set(curScriptID, false)
	cyclicItems.Delete(curScriptID)

	return false
}

func ValidateScripts(scrpts tasks.Scripts) error {
	sciptIDToNodesMap := make(map[string][]string)
	errs := utils.Errors{}

	for _, script := range scrpts {
		sciptIDToNodesMap[script.ID] = make([]string, 0)
		for _, task := range script.Tasks {
			for _, reqName := range task.GetRequirements() {
				sciptIDToNodesMap[script.ID] = append(sciptIDToNodesMap[script.ID], reqName)

				if reqName == script.ID {
					errs.Add(fmt.Errorf("task at path '%s' cannot require own script '%s'", task.GetPath(), script.ID))
					continue
				}
			}
		}
	}

	requestStack := orderedmap.NewOrderedMap()
	visited := make(map[string]bool)
	for curScriptID := range sciptIDToNodesMap {
		cyclicItms := orderedmap.NewOrderedMap()
		isCyclic := isCyclic(curScriptID, sciptIDToNodesMap, visited, requestStack, cyclicItms)
		if isCyclic {
			errs.Add(fmt.Errorf("cyclic requirements are detected: '%s'", cyclicItms.Keys()))
			break
		}
	}

	return errs.ToError()
}
