package exec

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSApi(t *testing.T) {
	cmd := exec.Command("echo", "123")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	cmdRunner := OSApi{}
	err := cmdRunner.Run(cmd)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, "123\n", outBuf.String())
}
