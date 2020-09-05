package script

import (
	"context"

	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/utils"

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
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskBuilder{},
			tasks.FileManaged:    &tasks.FileManagedTaskBuilder{},
		}),
	}

	execRouter := tasks.ExecutorRouter{
		Executors: map[string]tasks.Executor{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskExecutor{
				Runner: exec.SystemRunner{
					SystemAPI: exec.OSApi{},
				},
				FsManager: &utils.FsManager{},
			},
			tasks.FileManaged: &tasks.FileManagedTaskExecutor{
				Runner: exec.SystemRunner{
					SystemAPI: exec.OSApi{},
				},
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			},
		},
	}

	scripts, err := parser.BuildScripts()

	if err != nil {
		return err
	}

	runner := Runner{
		ExecutorRouter: execRouter,
	}

	err = runner.Run(context.Background(), scripts)

	return err
}
