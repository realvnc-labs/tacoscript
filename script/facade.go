package script

import (
	"context"
	"fmt"

	"github.com/cloudradar-monitoring/tacoscript/parse"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

// RunScript main entry point for the script execution
func RunScript(scriptPath string) error {
	fileDataProvider := parse.FileDataProvider{
		Path: scriptPath,
	}

	parser := parse.Parser{
		DataProvider: fileDataProvider,
		TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskBuilder{
				Runner:               tasks.OSCmdRunner{},
				UserSystemInfoParser: tasks.OSUserSystemInfoParser{},
			},
		}),
	}

	scripts, err := parser.ParseScripts()

	if err != nil {
		return err
	}

	runner := Runner{}

	res := runner.Run(context.Background(), scripts)

	fmt.Printf("%+v", res)
	return nil
}
