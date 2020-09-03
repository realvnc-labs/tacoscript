package apptest

import (
	"github.com/cloudradar-monitoring/tacoscript/conv"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func BuildExpectedEnvs(expectedEnvs map[interface{}]interface{}) []interface{} {
	envs := make([]interface{}, 0, len(expectedEnvs))
	for envKey, envValue := range expectedEnvs {
		envs = append(envs, map[interface{}]interface{}{
			envKey: envValue,
		})
	}

	return envs
}

func AssertEnvValuesMatch(t *testing.T, expectedEnvs conv.KeyValues, actualCmdEnvs []string) {
	expectedRawEnvs := expectedEnvs.ToEqualSignStrings()
	notFoundEnvs := make([]string, 0, len(expectedEnvs))
	for _, expectedRawEnv := range expectedRawEnvs {
		foundEnv := false
		for _, actualCmdEnv := range actualCmdEnvs {
			if expectedRawEnv == actualCmdEnv {
				foundEnv = true
				break
			}
		}

		if !foundEnv {
			notFoundEnvs = append(notFoundEnvs, expectedRawEnv)
		}
	}

	assert.Empty(
		t,
		notFoundEnvs,
		"was not able to find expected environment variables %s in cmd envs %s",
		strings.Join(notFoundEnvs, ", "),
		strings.Join(actualCmdEnvs, ", "),
	)
}
