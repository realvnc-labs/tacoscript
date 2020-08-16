package tasks

import (
	"errors"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/stretchr/testify/assert"
)

type BuilderMock struct {
	TypeName     string
	Path         string
	Context      []map[string]interface{}
	TaskToReturn Task
	ErrToReturn  error
}

func (bm *BuilderMock) Build(typeName, path string, context []map[string]interface{}) (Task, error) {
	bm.TypeName = typeName
	bm.Path = path
	bm.Context = context

	return bm.TaskToReturn, bm.ErrToReturn
}

func TestBuildWithRouting(t *testing.T) {
	successBuilder := &BuilderMock{
		TaskToReturn: &CmdRunTask{TypeName: "successTask", Path: "someSuccessPath", FsManager: &utils.FsManagerMock{}},
		ErrToReturn:  nil,
	}

	failBuilder := &BuilderMock{
		TaskToReturn: &CmdRunTask{TypeName: "failedTask", Path: "someFailedPath", FsManager: &utils.FsManagerMock{}},
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

	assert.Equal(t, "successTask", task.GetName())
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
