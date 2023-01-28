package tasks

import (
	"fmt"
	"strings"

	"github.com/cloudradar-monitoring/tacoscript/conv"
	"github.com/cloudradar-monitoring/tacoscript/utils"
	"gopkg.in/yaml.v2"
)

type Builder interface {
	Build(typeName, path string, context interface{}) (Task, error)
}

type BuildRouter struct {
	Builders map[string]Builder
}

type taskParamMapFn func(t Task, path string, val interface{}) error

type taskParamsFnMap map[string]taskParamMapFn

func NewBuilderRouter(builders map[string]Builder) BuildRouter {
	return BuildRouter{
		Builders: builders,
	}
}

func (br BuildRouter) Build(typeName, path string, params interface{}) (Task, error) {
	builder, ok := br.Builders[typeName]
	if !ok {
		return nil, fmt.Errorf("no builders registered for task type '%s'", typeName)
	}

	return builder.Build(typeName, path, params)
}

func Build(typeName, path string, params interface{}, task Task, fnMap taskParamsFnMap) (errs *utils.Errors) {
	errs = &utils.Errors{}

	for _, item := range params.([]interface{}) {
		row := item.(yaml.MapSlice)[0]
		key := row.Key.(string)
		val := row.Value
		mapFn, ok := fnMap[key]
		if !ok {
			continue
		}
		errs.Add(mapFn(task, path, val))
	}

	return errs
}

func parseCreatesField(val interface{}, path string) (createsItems []string, err error) {
	createsItems = make([]string, 0)
	switch typedVal := val.(type) {
	case string:
		createsItems = append(createsItems, typedVal)
	default:
		createsItems, err = conv.ConvertToValues(val, path)
	}

	return
}

func parseRequireField(val interface{}, path string) (requireItems []string, err error) {
	requireItems = make([]string, 0)
	switch typedVal := val.(type) {
	case string:
		requireItems = append(requireItems, typedVal)
	default:
		requireItems, err = conv.ConvertToValues(val, path)
	}

	return
}

func parseOnlyIfField(val interface{}, path string) (onlyIf []string, err error) {
	onlyIf = make([]string, 0)
	switch typedVal := val.(type) {
	case string:
		onlyIf = append(onlyIf, typedVal)
	default:
		onlyIf, err = conv.ConvertToValues(val, path)
	}

	return onlyIf, err
}

func parseUnlessField(val interface{}, path string) (unless []string, err error) {
	unless = make([]string, 0)
	switch typedVal := val.(type) {
	case string:
		unless = append(unless, typedVal)
	default:
		unless, err = conv.ConvertToValues(val, path)
	}

	return unless, err
}

func parseBoolField(val interface{}) bool {
	boolStr := strings.TrimSpace(fmt.Sprint(val))
	switch boolStr {
	case "":
		return false
	case "0":
		return false
	case "false":
		return false
	default:
		return true
	}
}
