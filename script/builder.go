package script

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/tasks"

	yaml2 "gopkg.in/yaml.v2"
)

type FileDataProvider struct {
	Path string
}

func (fdp FileDataProvider) Read() ([]byte, error) {
	return ioutil.ReadFile(fdp.Path)
}

type RawDataProvider interface {
	Read() ([]byte, error)
}

type Builder struct {
	DataProvider RawDataProvider
	TaskBuilder  tasks.Builder
}

func (p Builder) BuildScripts() (tasks.Scripts, error) {
	yamlFile, err := p.DataProvider.Read()
	if err != nil {
		return tasks.Scripts{}, err
	}

	rawScripts := map[string]map[string][]map[string]interface{}{}
	err = yaml2.Unmarshal(yamlFile, &rawScripts)
	if err != nil {
		return tasks.Scripts{}, err
	}

	scripts := make(tasks.Scripts, 0, len(rawScripts))
	errs := utils.Errors{}
	for scriptID, rawTasks := range rawScripts {
		script := tasks.Script{
			ID:    scriptID,
			Tasks: make([]tasks.Task, 0, len(rawTasks)),
		}
		index := 0
		for taskTypeID, taskContext := range rawTasks {
			index++
			task, e := p.TaskBuilder.Build(taskTypeID, fmt.Sprintf("%s.%s[%d]", scriptID, taskTypeID, index), taskContext)
			if e != nil {
				return tasks.Scripts{}, e
			}

			errs.Add(task.Validate())
			script.Tasks = append(script.Tasks, task)
		}

		scripts = append(scripts, script)
	}
	err = ValidateScripts(scripts)
	errs.Add(err)

	return scripts, errs.ToError()
}
