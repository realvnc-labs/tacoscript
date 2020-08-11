package script

import "github.com/cloudradar-monitoring/tacoscript/tasks"

func ValidateScripts(scrpts tasks.Scripts) error {
	scriptRequiredByMap := make(map[string]string)
	for _, script := range scrpts {
		for _, task := range script.Tasks {
			for _, reqName := range task.GetRequirements() {
				scriptRequiredByMap[reqName] = script.ID
				_, ok := scriptRequiredByMap[script.ID]
				if ok {
					continue
				}
			}
		}
	}
	return nil
}
