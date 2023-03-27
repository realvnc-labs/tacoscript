package script

import (
	"database/sql"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	filemanagedbuilder "github.com/realvnc-labs/tacoscript/tasks/filemanaged/builder"
	"github.com/realvnc-labs/tacoscript/utils"
)

func TestTaskBuilderFromRawYaml(t *testing.T) {
	testCases := []struct {
		YamlInput      string
		expectedScript tasks.Script
		expectedError  string
	}{
		{
			YamlInput:     "",
			expectedError: "empty script provided: nothing to execute",
		},
		{
			YamlInput:     "kkk",
			expectedError: "invalid script provided",
		},
		{
			YamlInput: `
maintain-my-file:
  file.managed:
    - name: C:\temp\npp.7.8.8.Installer.x64.exe
    - source: https://github.com/notepad-plus-plus
    - source_hash: 79eef25f9b0b2c642c62b7f737d4f53f
    - makedirs: true # default false
    - replace: false # default true
    - skip_verify: true # default false
    - creates: 'C:\Program Files\notepad++\notepad++.exe'
`,
			expectedScript: tasks.Script{
				ID: "maintain-my-file",
				Tasks: []tasks.CoreTask{
					&filemanaged.FmTask{
						TypeName: filemanaged.TaskTypeFileManaged,
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
						User:       "",
						Group:      "",
						Encoding:   "",
						Mode:       0,
						OnlyIf:     nil,
						Require:    nil,
					},
				},
			},
		},
		{
			YamlInput: `
maintain-another-file:
  file.managed:
    - name: /tmp/my-file.txt
    - contents: |
       My file content
       goes here
       Funny file
    - user: root
    - group: www-data
    - mode: 0755
    - encoding: UTF-8
    - onlyif:
      - which apache2
      - grep -q foo /tmp/bla
`,
			expectedScript: tasks.Script{
				ID: "maintain-another-file",
				Tasks: []tasks.CoreTask{
					&filemanaged.FmTask{
						TypeName: filemanaged.TaskTypeFileManaged,
						Path:     "maintain-another-file.file.managed[1]",
						Name:     "/tmp/my-file.txt",
						Contents: sql.NullString{
							Valid: true,
							String: `My file content
goes here
Funny file
`,
						},
						MakeDirs:   false,
						Replace:    true,
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
	}

	for _, testCase := range testCases {
		dataProviderMock := RawDataProviderMock{DataToReturn: testCase.YamlInput}
		parser := Builder{
			DataProvider: dataProviderMock,
			TaskBuilder: builder.NewBuilderRouter(map[string]builder.Builder{
				filemanaged.TaskTypeFileManaged: filemanagedbuilder.FileManagedTaskBuilder{},
			}),
			TemplateVariablesProvider: TemplateVariablesProviderMock{},
		}

		scripts, err := parser.BuildScripts()
		if testCase.expectedError != "" {
			require.Contains(t, err.Error(), testCase.expectedError)
			continue
		}

		require.NoError(t, err)

		assert.True(t, cmp.Equal(tasks.Scripts{testCase.expectedScript}, scripts, cmpopts.IgnoreUnexported(filemanaged.FmTask{})))
	}
}
