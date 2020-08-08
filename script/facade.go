package script

import (
	"context"
	"fmt"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

//RunScript main entry point for the script execution
func RunScript(scriptPath string) error {
	fileDataProvider := tasks.FileDataProvider{
		Path: scriptPath,
	}

	parser := tasks.Parser{
		DataProvider: fileDataProvider,
		TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskBuilder{
				Runner: tasks.OSCmdRunner{},
			},
		}),
	}

	scripts, err := parser.ParseScripts(fileDataProvider)

	if err != nil {
		return err
	}

	runner := Runner{}

	res := runner.Run(context.Background(), scripts)

	fmt.Printf("%+v", res)
	return nil
}
