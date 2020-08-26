package tasks

import (
	"fmt"

	"github.com/cloudradar-monitoring/tacoscript/conv"
	"github.com/cloudradar-monitoring/tacoscript/utils"
)

type Builder interface {
	Build(typeName, path string, context []map[string]interface{}) (Task, error)
}

type BuildRouter struct {
	Builders map[string]Builder
}

func NewBuilderRouter(builders map[string]Builder) BuildRouter {
	return BuildRouter{
		Builders: builders,
	}
}

func (br BuildRouter) Build(typeName, path string, context []map[string]interface{}) (Task, error) {
	builder, ok := br.Builders[typeName]
	if !ok {
		return nil, fmt.Errorf("no builders registered for task type '%s'", typeName)
	}

	return builder.Build(typeName, path, context)
}

func parseCreatesField(val interface{}, path string, errs *utils.Errors) (createsItems []string) {
	createsItems = make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		createsItems = append(createsItems, typedVal)
	default:
		createsItems, err = conv.ConvertToValues(val, path)
		errs.Add(err)
	}

	return
}

func parseRequireField(val interface{}, path string, errs *utils.Errors) (requireItems []string) {
	requireItems = make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		requireItems = append(requireItems, typedVal)
	default:
		requireItems, err = conv.ConvertToValues(val, path)
		errs.Add(err)
	}

	return
}

func parseOnlyIfField(val interface{}, path string, errs *utils.Errors) (onlyIf []string) {
	onlyIf = make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		onlyIf = append(onlyIf, typedVal)
	default:
		onlyIf, err = conv.ConvertToValues(val, path)
		errs.Add(err)
	}

	return
}

func parseUnlessField(val interface{}, path string, errs *utils.Errors) (unless []string) {
	unless = make([]string, 0)
	var err error
	switch typedVal := val.(type) {
	case string:
		unless = append(unless, typedVal)
	default:
		unless, err = conv.ConvertToValues(val, path)
		errs.Add(err)
	}

	return
}
