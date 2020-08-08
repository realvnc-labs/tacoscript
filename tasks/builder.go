package tasks

import "fmt"

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
