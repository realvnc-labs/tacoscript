package script

import (
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
	"github.com/cloudradar-monitoring/tacoscript/utils"
	"github.com/stretchr/testify/assert"
)

func TestTaskBuilderFromRawYaml(t *testing.T) {
	testCases := []struct {
		YamlInput       string
		ExpectedErrMsg  string
		expectedScripts tasks.Scripts
	}{
		{
			YamlInput: `maintain-my-file:
  file.managed:
    - name: C:\temp\npp.7.8.8.Installer.x64.exe
    - source: https://github.com/notepad-plus-plus
    - source_hash: 79eef25f9b0b2c642c62b7f737d4f53f
    - makedirs: true # default false
    - replace: false # default true
    - skip_verify: true # default false
    - creates: 'C:\Program Files\notepad++\notepad++.exe'
`,
			expectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "maintain-my-file",
					Tasks: []tasks.Task{
						&tasks.FileManagedTask{
							TypeName:   tasks.FileManaged,
							Path:       "maintain-my-file.file.managed[1]",
							Name:       "C:\\temp\\npp.7.8.8.Installer.x64.exe",
							Source:     "https://github.com/notepad-plus-plus",
							SourceHash: "79eef25f9b0b2c642c62b7f737d4f53f",
							MakeDirs:   true,
							Replace:    false,
							SkipVerify: true,
							Creates:    []string{"C:\\Program Files\\notepad++\\notepad++.exe"},
							Contents:   "",
							User:       "",
							Group:      "",
							Encoding:   "",
							Mode:       "",
							OnlyIf:     nil,
							Runner:     nil,
							FsManager:  nil,
							Require:    nil,
							Errors:     &utils.Errors{},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		dataProviderMock := RawDataProviderMock{DataToReturn: testCase.YamlInput}
		parser := Builder{
			DataProvider: dataProviderMock,
			TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
				tasks.FileManaged: tasks.FileManagedTaskBuilder{},
			}),
		}

		scripts, err := parser.BuildScripts()
		if testCase.ExpectedErrMsg != "" {
			assert.EqualError(t, err, testCase.ExpectedErrMsg, testCase.ExpectedErrMsg)
			continue
		}

		assert.EqualValues(t, testCase.expectedScripts, scripts)
	}
}
