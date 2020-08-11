package appos

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	cmd := exec.Command("echo", "123")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	cmdRunner := OsExecutor{}
	err := cmdRunner.Run(cmd)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, "123\n", outBuf.String())
}
