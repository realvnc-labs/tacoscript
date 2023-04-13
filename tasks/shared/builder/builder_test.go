package builder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
)

type BuilderMock struct {
	TypeName     string
	Path         string
	Context      interface{}
	TaskToReturn tasks.CoreTask
	ErrToReturn  error
}

func (bm *BuilderMock) Build(typeName, path string, ctx interface{}) (tasks.CoreTask, error) {
	bm.TypeName = typeName
	bm.Path = path
	bm.Context = ctx

	return bm.TaskToReturn, bm.ErrToReturn
}

func TestBuildWithRouting(t *testing.T) {
	successBuilder := &BuilderMock{
		TaskToReturn: &cmdrun.Task{TypeName: "successTask", Path: "someSuccessPath"},
		ErrToReturn:  nil,
	}

	failBuilder := &BuilderMock{
		TaskToReturn: &cmdrun.Task{TypeName: "failedTask", Path: "someFailedPath"},
		ErrToReturn:  errors.New("some error"),
	}

	br := BuildRouter{
		Builders: map[string]Builder{
			"successTask": successBuilder,
			"failedTask":  failBuilder,
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

	assert.Equal(t, "successTask", task.GetTypeName())
	assert.Equal(t, "someSuccessPath", task.GetPath())
	assert.Equal(t, "successTask", successBuilder.TypeName)
	assert.Equal(t, "someSuccessPath", successBuilder.Path)
	assert.Equal(t, ctx, successBuilder.Context)

	_, err2 := br.Build("failedTask", "someFailedPath", ctx)
	assert.EqualError(t, err2, "some error")
	if err2 == nil {
		return
	}
	assert.Equal(t, "failedTask", failBuilder.TypeName)
	assert.Equal(t, "someFailedPath", failBuilder.Path)
	assert.Equal(t, ctx, failBuilder.Context)
}
