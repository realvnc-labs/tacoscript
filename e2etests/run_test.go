package e2etests

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/script"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var testTacoScript = flag.String("script", "", "Single Tacoscript to test")

type expectedSummary struct {
	Succeeded     int `yaml:"Succeeded"`
	Failed        int `yaml:"Failed"`
	Aborted       int `yaml:"Aborted"`
	Changes       int `yaml:"Changes"`
	TotalTasksRun int `yaml:"TotalTasksRun"`
}

type expectedTaskResult struct {
	ID              string   `yaml:"ID"`
	Result          bool     `yaml:"Result"`
	ChangesContains []string `yaml:"ChangesContains"`
	CommentContains []string `yaml:"CommentContains"`
	HasChanges      bool     `yaml:"HasChanges"`
}

type expected struct {
	Summary     expectedSummary      `yaml:"Summary"`
	TaskResults []expectedTaskResult `yaml:"TaskResults"`
	PreExec     string               `yaml:"PreExec"`
	PostExec    string               `yaml:"PostExec"`
}

type testYml struct {
	Run    yaml.MapSlice `yaml:"Run"`
	On     []string      `yaml:"On"`
	OnlyIf string        `yaml:"OnlyIf"`
	Expect expected      `yaml:"Expect"`
}

func flatten(in map[string]interface{}) string {
	var response string
	for k, v := range in {
		response += fmt.Sprintf("%s: %s ", k, v)
	}
	return response
}

func isIn(in string, list []string) bool {
	for _, s := range list {
		if s == in {
			return true
		}
	}
	return false
}

func runCmd(command string) (stdout []byte, err error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd = exec.Command("sh", "-c", command)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", command)
	}

	stdout, err = cmd.Output()
	return stdout, err
}

func TestTacoScript(t *testing.T) {
	files, err := os.ReadDir(".")
	assert.NoError(t, err)
	if *testTacoScript != "" {
		t.Logf("Running single tacoscript '%s'", *testTacoScript)
		_, err := os.Stat(*testTacoScript + ".yaml")
		assert.NoError(t, err)
	}
	for _, file := range files {
		if *testTacoScript != "" && file.Name() != (*testTacoScript+".yaml") {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		inFile := file.Name()
		testScript, err := os.ReadFile(inFile)
		assert.NoError(t, err)

		var test testYml
		err = yaml.Unmarshal(testScript, &test)
		require.NoError(t, err, "Error unmarshalling "+inFile)

		tacoTempFile := t.TempDir() + "/" + inFile
		run, err := yaml.Marshal(test.Run)
		require.NoError(t, err)
		if !isIn(runtime.GOOS, test.On) {
			t.Logf("Skipping test %s not runable on %s, runs on %s only", inFile, runtime.GOOS, test.On)
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			if test.OnlyIf != "" {
				if _, err = runCmd(test.OnlyIf); err != nil {
					// Skip test because condition not met
					t.Logf("Onlyif command failed. Test skipped.")
					return
				}
			}

			err = os.WriteFile(tacoTempFile, run, 0600) // Write tacoscript to a temp file
			require.NoError(t, err)

			// Execute a command before running the tacoscript
			if test.Expect.PreExec != "" {
				var stdout []byte
				stdout, err = runCmd(test.Expect.PreExec)
				require.NoError(t, err, fmt.Sprintf("Executing PreExec command '%s': %s %s", test.Expect.PreExec, stdout, err))
			}

			// Run the tacoscript and capture the output
			t.Logf("Running tacoscript %s", inFile)
			var output bytes.Buffer
			err = script.RunScript(tacoTempFile, false, &output)
			require.NoError(t, err)

			// Execute a command after running the tacoscript
			if test.Expect.PostExec != "" {
				var stdout []byte
				stdout, err = runCmd(test.Expect.PostExec)
				require.NoError(t, err, fmt.Sprintf("Executing PostExec command '%s': %s %s", test.Expect.PostExec, stdout, err))
				t.Logf("PostExec script returns:\n%s", stdout)
			}

			// Map the captured output to a struct
			t.Logf("Captured script output:\n%s", output.String())
			result := script.Result{}
			err = yaml.Unmarshal(output.Bytes(), &result)

			// Validate the results are as expected
			assert.NoError(t, err)
			assert.Equal(t, tacoTempFile, result.Summary.Script, "Validating Summary.Script")
			assert.Equal(t, test.Expect.Summary.Changes, result.Summary.Changes, "Validating Summary.Changes")
			assert.Equal(t, test.Expect.Summary.Succeeded, result.Summary.Succeeded, "Validating Summary.Succeeded")
			assert.Equal(t, test.Expect.Summary.Aborted, result.Summary.Aborted, "Validating Summary.Aborted")
			assert.Equal(t, test.Expect.Summary.Failed, result.Summary.Failed, "Validating Summary.Failed")
			assert.Equal(t, test.Expect.Summary.TotalTasksRun, result.Summary.TotalTasksRun, "Validating Summary.TotalTasksRun")

			// Range over the results of each task
			for _, taskResult := range result.Results {
				t.Logf("Validating results of task '%s'", taskResult.ID)
				var tested = false
				for _, et := range test.Expect.TaskResults {
					if et.ID == taskResult.ID {
						tested = true
						if len(et.ChangesContains) > 0 {
							et.HasChanges = true
							for _, change := range et.ChangesContains {
								assert.Contains(t, flatten(taskResult.Changes), change, "Validating changes")
							}
							assert.True(t, et.HasChanges, "Validating has changes")
						} else {
							assert.False(t, et.HasChanges, "Validating has no changes")
						}
						if len(et.CommentContains) > 0 {
							for _, cc := range et.CommentContains {
								assert.Contains(t, taskResult.Comment, cc, "Validating comment")
							}
						}
					}
				}
				assert.True(t, tested, fmt.Sprintf("Task '%s' has no expected result defined.", taskResult.ID))
			}
		})
	}
}
