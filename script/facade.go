package script

import (
	"context"
	"fmt"

	"github.com/cloudradar-monitoring/tacoscript/appos"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

// RunScript main entry point for the script execution
func RunScript(scriptPath string) error {
	fileDataProvider := FileDataProvider{
		Path: scriptPath,
	}

	parser := Builder{
		DataProvider: fileDataProvider,
		TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskBuilder{
				OsExecutor: appos.OsExecutor{},
			},
		}),
	}

	scripts, err := parser.BuildScripts()

	if err != nil {
		return err
	}

	runner := Runner{}

	res := runner.Run(context.Background(), scripts)

	fmt.Printf("%+v", res)
	return nil
}
