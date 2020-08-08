package tasks

import (
	"fmt"
	"io/ioutil"

	"github.com/goccy/go-yaml"
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

func (p Parser) ParseScripts(dataProvider RawDataProvider) (Scripts, error) {
	yamlFile, err := dataProvider.Read()
	if err != nil {
		return Scripts{}, err
	}

	rawScripts := map[string]map[string][]map[string]interface{}{}
	err = yaml.Unmarshal(yamlFile, &rawScripts)
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

func extractEnvFields(envs interface{}, path string) ([]Env, error) {
	rawEnvs, ok := envs.([]interface{})
	if !ok {
		return []Env{}, fmt.Errorf("wrong env variables value: array is exected at path %s.%s but got %v", path, EnvField, envs)
	}

	res := make([]Env, 0, len(rawEnvs))

	for _, rawEnv := range rawEnvs {
		envMap, ok := rawEnv.(map[string]interface{})
		if !ok {
			return []Env{}, fmt.Errorf("wrong env variables value: array of scalar values is exected at path %s.%s but got %v", path, EnvField, envs)
		}

		for envKey, envVal := range envMap {
			res = append(res, Env{
				Key:   envKey,
				Value: fmt.Sprint(envVal),
			})
		}
	}

	return res, nil
}
