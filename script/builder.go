package script

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
)

type FileDataProvider struct {
	Path string
}

func (fdp FileDataProvider) Read() ([]byte, error) {
	return os.ReadFile(fdp.Path)
}

type RawDataProvider interface {
	Read() ([]byte, error)
}

type TemplateVariablesProvider interface {
	GetTemplateVariables() (utils.TemplateVarsMap, error)
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
	err = yaml.Unmarshal(yamlBody, &rawScripts)
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
		if steps, ok := rawTask.Value.(yaml.MapSlice); ok {
			for _, step := range steps {
				taskTypeID := step.Key.(string)
				taskParams := step.Value
				index++

				task, err := p.TaskBuilder.Build(taskTypeID, fmt.Sprintf("%s.%s[%d]", scriptID, taskTypeID, index), taskParams)
				if err != nil {
					return tasks.Scripts{}, err
				}

				err = task.Validate(runtime.GOOS)
				if err != nil {
					errs.Add(err)
				}

				script.Tasks = append(script.Tasks, task)
			}
		} else {
			errs.Add(fmt.Errorf("script failed to run. input YAML is malformed"))
		}
		scripts = append(scripts, script)
	}
	err = ValidateScripts(scripts)
	errs.Add(err)

	return scripts, errs.ToError()
}

func (p Builder) render(templateData []byte, variables utils.TemplateVarsMap) (result []byte, err error) {
	templ := template.New("goyaml")

	pageTemplate, err := templ.Option("missingkey=zero").Parse(string(templateData))
	if err != nil {
		return result, err
	}

	buf := bytes.Buffer{}

	err = pageTemplate.Execute(&buf, variables)

	return buf.Bytes(), err
}
