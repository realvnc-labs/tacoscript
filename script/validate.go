package script

import (
	"fmt"
	"strings"

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
	scriptIDToNodesMap := make(map[string][]string)
	scriptIDsMap := make(map[string]bool, len(scrpts))
	requirements := make(map[string]string)
	errs := utils.Errors{}

	for _, script := range scrpts {
		scriptIDToNodesMap[script.ID] = make([]string, 0)
		scriptIDsMap[script.ID] = true
		for _, task := range script.Tasks {
			for k, reqName := range task.GetRequirements() {
				requirements[reqName] = fmt.Sprintf("%s.%s[%d]", task.GetPath(), tasks.RequireField, k)
				scriptIDToNodesMap[script.ID] = append(scriptIDToNodesMap[script.ID], reqName)

				if reqName == script.ID {
					errs.Add(fmt.Errorf("task at path '%s' cannot require own script '%s'", task.GetPath(), script.ID))
					continue
				}
			}
		}
	}

	reqFailures := make([]string, 0, len(requirements))
	for reqName, reqPath := range requirements {
		if _, ok := scriptIDsMap[reqName]; !ok {
			reqFailures = append(reqFailures, fmt.Sprintf("'%s' at path '%s'", reqName, reqPath))
		}
	}

	if len(reqFailures) > 0 {
		errs.Add(fmt.Errorf("missing required scripts %s", strings.Join(reqFailures, ", ")))
	}

	requestStack := orderedmap.NewOrderedMap()
	visited := make(map[string]bool)
	for curScriptID := range scriptIDToNodesMap {
		cyclicItms := orderedmap.NewOrderedMap()
		isCyclic := isCyclic(curScriptID, scriptIDToNodesMap, visited, requestStack, cyclicItms)
		if isCyclic {
			errs.Add(fmt.Errorf("cyclic requirements are detected: '%s'", cyclicItms.Keys()))
			break
		}
	}

	return errs.ToError()
}
