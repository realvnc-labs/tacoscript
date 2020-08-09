package io

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncWriter(t *testing.T) {
	buf := new(bytes.Buffer)

	tf := FuncWriter{
		Callback: buf.Write,
	}

	dataToWrite := []byte("some data")
	writtenCount, err := tf.Write(dataToWrite)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, len(dataToWrite), writtenCount)
	assert.Equal(t, string(dataToWrite), buf.String())

	tf2 := FuncWriter{
		Callback: func(p []byte) (n int, err error) {
			return 0, errors.New("some error")
		},
	}
	_, err2 := tf2.Write(dataToWrite)
	assert.EqualError(t, err2, "some error")
}
