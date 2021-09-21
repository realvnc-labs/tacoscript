package script

import (
	"context"

	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/pkg"
	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

// RunScript main entry point for the script execution
func RunScript(scriptPath string, abortOnError bool) error {
	fileDataProvider := FileDataProvider{
		Path: scriptPath,
	}

	parser := Builder{
		DataProvider: fileDataProvider,
		TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskBuilder{},
			tasks.FileManaged:    &tasks.FileManagedTaskBuilder{},
			tasks.PkgInstalled:   &tasks.PkgTaskBuilder{},
			tasks.PkgRemoved:     &tasks.PkgTaskBuilder{},
			tasks.PkgUpgraded:    &tasks.PkgTaskBuilder{},
		}),
		TemplateVariablesProvider: utils.OSDataProvider{},
	}

	cmdRunner := exec.SystemRunner{
		SystemAPI: exec.OSApi{},
	}

	pkgTaskManager := pkg.PackageTaskManager{
		Runner:                          cmdRunner,
		ManagementCmdsProviderBuildFunc: pkg.BuildManagementCmdsProviders,
	}
	pkgTaskExecutor := &tasks.PkgTaskExecutor{
		PackageManager: pkgTaskManager,
		Runner:         cmdRunner,
	}
	execRouter := tasks.ExecutorRouter{
		Executors: map[string]tasks.Executor{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskExecutor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			tasks.FileManaged: &tasks.FileManagedTaskExecutor{
				Runner:      cmdRunner,
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			},
			tasks.PkgInstalled: pkgTaskExecutor,
			tasks.PkgRemoved:   pkgTaskExecutor,
			tasks.PkgUpgraded:  pkgTaskExecutor,
		},
	}

	scripts, err := parser.BuildScripts()

	if err != nil {
		return err
	}

	runner := Runner{
		DataProvider:   fileDataProvider,
		ExecutorRouter: execRouter,
	}

	err = runner.Run(context.Background(), scripts, abortOnError)

	return err
}
