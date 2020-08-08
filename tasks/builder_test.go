package tasks

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type BuilderMock struct {
	TypeName string
	Path string
	Context []map[string]interface{}
	TaskToReturn Task
	ErrToReturn error
}

func (bm *BuilderMock) Build(typeName, path string, context []map[string]interface{}) (Task, error) {
	bm.TypeName = typeName
	bm.Path = path
	bm.Context = context

	return bm.TaskToReturn, bm.ErrToReturn
}

func TestBuildWithRouting(t *testing.T) {
	successBuilder := &BuilderMock{
		TaskToReturn: &CmdRunTask{TypeName:  "successTask", Path: "someSuccessPath"},
		ErrToReturn:  nil,
	}

	failBuilder := &BuilderMock{
		TaskToReturn: &CmdRunTask{TypeName:  "failedTask", Path: "someFailedPath"},
		ErrToReturn:  errors.New("some error"),
	}

	br := BuildRouter{
		Builders: map[string]Builder{
			"successTask": successBuilder,
			"failedTask": failBuilder,
		},
	}

	ctx := []map[string]interface{}{
		{
			"someKey": "someValue",
		},
	}

	task, err := br.Build("successTask", "someSuccessPath", ctx)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, "successTask",  task.GetName())
	assert.Equal(t, "someSuccessPath",  task.GetPath())
	assert.Equal(t, "successTask", successBuilder.TypeName)
	assert.Equal(t, "someSuccessPath", successBuilder.Path)
	assert.Equal(t, ctx, successBuilder.Context)
}
