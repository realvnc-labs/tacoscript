package apptest

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"strings"
	"testing"
)

func AssertCmdsPartiallyMatch(t *testing.T, expectedCmds []string, actualExecutedCmds []*exec.Cmd) {
	notFoundCmds := make([]string, 0, len(expectedCmds))

	executedCmdStrs := make([]string, 0, len(actualExecutedCmds))
	for _, actualCmd := range actualExecutedCmds {
		executedCmdStrs = append(executedCmdStrs, actualCmd.String())
	}

	for _, expectedCmdStr := range expectedCmds {
		cmdFound := false
		for _, actualCmdStr := range executedCmdStrs {
			if strings.HasSuffix(actualCmdStr, expectedCmdStr) {
				cmdFound = true
				break
			}
		}
		if !cmdFound {
			notFoundCmds = append(notFoundCmds, expectedCmdStr)
		}
	}

	assert.Empty(
		t,
		notFoundCmds,
		"was not able to find following expected commands '%s' in the list of executed commands '%s'",
		strings.Join(notFoundCmds, ", "),
		strings.Join(executedCmdStrs, ", "),
	)
}
