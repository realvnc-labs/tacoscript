package tasks

import (
	"fmt"
	"io/ioutil"

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

type Parser struct {
	DataProvider RawDataProvider
	TaskBuilder  Builder
}

func (p Parser) ParseScripts() (Scripts, error) {
	yamlFile, err := p.DataProvider.Read()
	if err != nil {
		return Scripts{}, err
	}

	rawScripts := map[string]map[string][]map[string]interface{}{}
	err = yaml2.Unmarshal(yamlFile, &rawScripts)
	if err != nil {
		return Scripts{}, err
	}

	scripts := make(Scripts, 0, len(rawScripts))
	errs := ValidationErrors{}
	for scriptID, rawTasks := range rawScripts {
		script := Script{
			ID:    scriptID,
			Tasks: make([]Task, 0, len(rawTasks)),
		}
		index := 0
		for taskTypeID, taskContext := range rawTasks {
			index++
			task, err := p.TaskBuilder.Build(taskTypeID, fmt.Sprintf("%s.%s[%d]", scriptID, taskTypeID, index), taskContext)
			if err != nil {
				return Scripts{}, err
			}

			errs.Add(task.Validate())
			script.Tasks = append(script.Tasks, task)
		}

		scripts = append(scripts, script)
	}

	return scripts, errs.ToError()
}
