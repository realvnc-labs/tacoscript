package script

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
)

type RequirementsTaskMock struct {
	RequirementsToGive []string
	Path               string
	OnlyIf             []string
	Unless             []string
	Creates            []string
}

func (rtm *RequirementsTaskMock) GetTypeName() string {
	return ""
}

func (rtm *RequirementsTaskMock) Execute(ctx context.Context) executionresult.ExecutionResult {
	return executionresult.ExecutionResult{}
}

func (rtm *RequirementsTaskMock) Validate(goos string) error {
	return nil
}

func (rtm *RequirementsTaskMock) GetPath() string {
	return rtm.Path
}

func (rtm *RequirementsTaskMock) GetRequirements() []string {
	return rtm.RequirementsToGive
}

func (rtm *RequirementsTaskMock) GetOnlyIfCmds() []string {
	return rtm.OnlyIf
}

func (rtm *RequirementsTaskMock) GetUnlessCmds() []string {
	return rtm.Unless
}

func (rtm *RequirementsTaskMock) GetCreatesFilesList() []string {
	return rtm.Creates
}

type errorExpectation struct {
	messagePrefix  string
	availableParts []string
}

func TestScriptsValidation(t *testing.T) {
	testCases := []struct {
		name             string
		scripts          tasks.Scripts
		errorExpectation errorExpectation
	}{
		{
			name: "cycle of 2 scripts requiring each other",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script one",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script two"},
							Path:               "script one.RequirementsTaskMock[0]",
						},
					},
				},
				tasks.Script{
					ID: "script two",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script one"},
							Path:               "script two.RequirementsTaskMock[0]",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script one",
					"script two",
				},
			},
		},
		{
			name: "no requirement cycle",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 3",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{},
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 4"},
						},
					},
				},
				tasks.Script{
					ID:    "script 4",
					Tasks: []tasks.Task{},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "circling cycle",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 6",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 7"},
						},
					},
				},
				tasks.Script{
					ID: "script 7",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 8"},
						},
					},
				},
				tasks.Script{
					ID: "script 8",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
				tasks.Script{
					ID: "script 9",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script 6",
					"script 7",
					"script 8",
				},
			},
		},
		{
			name: "many_scripts_no_cycle",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 10",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 11"},
						},
					},
				},
				tasks.Script{
					ID: "script 11",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 12", "script 13"},
						},
					},
				},
				tasks.Script{
					ID: "script 13",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{},
						},
					},
				},
				tasks.Script{
					ID: "script 12",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{},
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "require itself",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 20",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 20"},
							Path:               "script 20 task 1 path 1",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "task at path 'script 20 task 1 path 1' cannot require own script 'script 20'",
			},
		},
		{
			name: "multiple required scripts not found",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 30",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 31", "script 32"},
							Path:               "path 1",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "missing required scripts",
				availableParts: []string{
					"'script 31' at path 'path 1.require[0]'",
					"'script 32' at path 'path 1.require[1]'",
				},
			},
		},
		{
			name: "all required scripts are found",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 32",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 33", "script 34"},
						},
					},
				},
				tasks.Script{
					ID:    "script 33",
					Tasks: []tasks.Task{&RequirementsTaskMock{}},
				},
				tasks.Script{
					ID:    "script 34",
					Tasks: []tasks.Task{&RequirementsTaskMock{}},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "some required scripts not found",
			scripts: tasks.Scripts{
				tasks.Script{
					ID: "script 35",
					Tasks: []tasks.Task{
						&RequirementsTaskMock{
							RequirementsToGive: []string{"script 36"},
						},
					},
				},
				tasks.Script{
					ID: "script 36",
					Tasks: []tasks.Task{&RequirementsTaskMock{
						RequirementsToGive: []string{"script 37"},
						Path:               "path 36",
					}},
				},
				tasks.Script{
					ID: "script 38",
					Tasks: []tasks.Task{&RequirementsTaskMock{
						RequirementsToGive: []string{"script 40"},
						Path:               "path 38",
					}},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "missing required scripts",
				availableParts: []string{
					"'script 37' at path 'path 36.require[0]'",
					"'script 40' at path 'path 38.require[0]'",
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateScripts(tc.scripts)
			if tc.errorExpectation.messagePrefix == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err == nil {
					return
				}

				assert.Contains(t, err.Error(), tc.errorExpectation.messagePrefix)

				for _, expectedMsgPart := range tc.errorExpectation.availableParts {
					assert.Contains(t, err.Error(), expectedMsgPart)
				}
			}
		})
	}
}
