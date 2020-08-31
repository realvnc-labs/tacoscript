package script

import (
	"net/url"
	"os"
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
maintain-another-file:
  file.managed:
    - name: /tmp/my-file.txt
    - contents: ddd
    - user: root
    - group: www-data
    - mode: 0755
    - encoding: UTF-8
    - onlyif:
      - which apache2
      - grep -q foo /tmp/bla
`,
			expectedScripts: tasks.Scripts{
				tasks.Script{
					ID: "maintain-my-file",
					Tasks: []tasks.Task{
						&tasks.FileManagedTask{
							TypeName: tasks.FileManaged,
							Path:     "maintain-my-file.file.managed[1]",
							Name:     "C:\\temp\\npp.7.8.8.Installer.x64.exe",
							Source: utils.Location{
								IsURL: true,
								URL: &url.URL{
									Scheme: "https",
									Host:   "github.com",
									Path:   "/notepad-plus-plus",
								},
								RawLocation: "https://github.com/notepad-plus-plus",
							},
							SourceHash: "79eef25f9b0b2c642c62b7f737d4f53f",
							MakeDirs:   true,
							Replace:    false,
							SkipVerify: true,
							Creates:    []string{"C:\\Program Files\\notepad++\\notepad++.exe"},
							Contents:   "",
							User:       "",
							Group:      "",
							Encoding:   "",
							Mode:       0,
							OnlyIf:     nil,
							Require:    nil,
						},
					},
				},
				tasks.Script{
					ID: "maintain-another-file",
					Tasks: []tasks.Task{
						&tasks.FileManagedTask{
							TypeName:   tasks.FileManaged,
							Path:       "maintain-another-file.file.managed[1]",
							Name:       "/tmp/my-file.txt",
							Contents:   `ddd`,
							MakeDirs:   false,
							Replace:    false,
							SkipVerify: false,
							User:       "root",
							Group:      "www-data",
							Encoding:   "UTF-8",
							Mode:       os.FileMode(0755),
							OnlyIf:     []string{"which apache2", "grep -q foo /tmp/bla"},
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

		assert.NoError(t, err)

		assert.EqualValues(t, testCase.expectedScripts, scripts)
	}
}
