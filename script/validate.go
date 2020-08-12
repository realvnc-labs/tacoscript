package script

import (
	"errors"
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/cloudradar-monitoring/tacoscript/utils"
)

type node struct {
	task           tasks.Task
	sourceScriptID string
	targetScriptID string
}

func isCyclic(curScriptID string, scriptIDToNodesMap map[string][]node, visited map[string]bool, requestStack map[string]bool) bool {
	if isInRequestStack, ok := requestStack[curScriptID]; ok && isInRequestStack {
		return true
	}

	if isVisited, ok := visited[curScriptID]; ok && isVisited {
		return false
	}

	visited[curScriptID] = true
	requestStack[curScriptID] = true

	if childNodes, ok := scriptIDToNodesMap[curScriptID]; ok {
		for _, childNode := range childNodes {
			isCyclic := isCyclic(childNode.targetScriptID, scriptIDToNodesMap, visited, requestStack)
			if isCyclic {
				return true
			}
		}
	}

	requestStack[curScriptID] = false

	return false
}

func ValidateScripts(scrpts tasks.Scripts) error {
	sciptIDToNodesMap := make(map[string][]node)
	errs := utils.Errors{}

	for _, script := range scrpts {
		sciptIDToNodesMap[script.ID] = make([]node, 0)
		for _, task := range script.Tasks {
			for _, reqName := range task.GetRequirements() {
				sciptIDToNodesMap[script.ID] = append(sciptIDToNodesMap[script.ID], node{
					task:           task,
					sourceScriptID: script.ID,
					targetScriptID: reqName,
				})

				if reqName == script.ID {
					errs.Add(fmt.Errorf("task at path '%s' cannot require own script %s", task.GetPath(), script.ID))
					continue
				}
			}
		}
	}

	visited, requestStack := make(map[string]bool, 0), make(map[string]bool, 0)
	isCyclicAll := false
	for curScriptID, _ := range sciptIDToNodesMap {
		isCyclic := isCyclic(curScriptID, sciptIDToNodesMap, visited, requestStack)
		if isCyclic {
			isCyclicAll = true
		}

		//for _, curNode := range curNodes {
		//	currentNodeSource := curNode.sourceScriptID
		//	nodesRequiringCurScript, ok := targetScriptIDToNodeMap[currentNodeSource]
		//	if !ok {
		//		continue
		//	}
		//
		//	for _, nodeRequiringCurScript := range nodesRequiringCurScript {
		//		if nodeRequiringCurScript.sourceScriptID == targetScriptID {
		//			errs.Add(fmt.Errorf("cyclic requirement detected: the task at '%s' of script %s"+
		//				"requires '%s' which has task at '%s' requiring script '%s'",
		//				curNode.task.GetPath(),
		//				currentNodeSource,
		//				curNode.targetScriptID,
		//				nodeRequiringCurScript.task.GetPath(),
		//				targetScriptID,
		//			))
		//		}
		//	}
		//}
	}

	if isCyclicAll {
		return errors.New("is cyclic")
	}
	return nil
}
