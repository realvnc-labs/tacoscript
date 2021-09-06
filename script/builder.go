package script

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/tasks"

	"gopkg.in/yaml.v2"
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

type TemplateVariablesProvider interface {
	GetTemplateVariables() (map[string]interface{}, error)
}

type Builder struct {
	DataProvider              RawDataProvider
	TaskBuilder               tasks.Builder
	TemplateVariablesProvider TemplateVariablesProvider
}

func (p Builder) BuildScripts() (tasks.Scripts, error) {
	yamlTemplate, err := p.DataProvider.Read()
	if err != nil {
		return tasks.Scripts{}, err
	}

	templateVariables, err := p.TemplateVariablesProvider.GetTemplateVariables()
	if err != nil {
		return tasks.Scripts{}, err
	}
	yamlBody, err := p.render(yamlTemplate, templateVariables)
	if err != nil {
		return tasks.Scripts{}, err
	}
	if len(yamlBody) == 0 {
		return tasks.Scripts{}, errors.New("empty script provided: nothing to execute")
	}

	rawScripts := yaml.MapSlice{}
	err = yaml2.Unmarshal(yamlBody, &rawScripts)
	if err != nil {
		return tasks.Scripts{}, fmt.Errorf("invalid script provided: %w", err)
	}

	scripts := make(tasks.Scripts, 0, len(rawScripts))
	errs := utils.Errors{}
	for _, rawTask := range rawScripts {
		scriptID := rawTask.Key.(string)
		script := tasks.Script{
			ID:    scriptID,
			Tasks: []tasks.Task{},
		}
		index := 0
		steps := rawTask.Value.(yaml.MapSlice)
		for _, step := range steps {
			taskTypeID := step.Key.(string)

			index++
			task, e := p.TaskBuilder.Build(taskTypeID, fmt.Sprintf("%s.%s[%d]", scriptID, taskTypeID, index), step.Value)
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

func (p Builder) render(templateData []byte, variables map[string]interface{}) (result []byte, err error) {
	templ := template.New("goyaml")

	pageTemplate, err := templ.Parse(string(templateData))
	if err != nil {
		return result, err
	}

	buf := bytes.Buffer{}

	err = pageTemplate.Execute(&buf, variables)

	return buf.Bytes(), err
}
