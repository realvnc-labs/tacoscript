package script

import (
	"context"
	"github.com/cloudradar-monitoring/tacoscript/pkg"
	"github.com/sirupsen/logrus"

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
			tasks.PkgInstalled:   &tasks.PkgTaskBuilder{},
			tasks.PkgRemoved:     &tasks.PkgTaskBuilder{},
			tasks.PkgUpgraded:    &tasks.PkgTaskBuilder{},
		}),
		TemplateVariablesProvider: utils.OSDataProvider{},
	}

	cmdRunner := exec.SystemRunner{
		SystemAPI: exec.OSApi{},
	}

	pkgCmdProviders, err := pkg.BuildManagementCmdsProviders()
	if err != nil {
		logrus.Warn(err.Error())
	}
	pkgTaskManager := pkg.PackageTaskManager{
		Runner:                     cmdRunner,
		PackageManagerCmdProviders: pkgCmdProviders,
	}
	pkgTaskExecutor := &tasks.PkgTaskExecutor{
		PackageManager: pkgTaskManager,
		Runner:         cmdRunner,
	}
	execRouter := tasks.ExecutorRouter{
		Executors: map[string]tasks.Executor{
			tasks.TaskTypeCmdRun: &tasks.CmdRunTaskExecutor{
				Runner: cmdRunner,
				FsManager: &utils.FsManager{},
			},
			tasks.FileManaged: &tasks.FileManagedTaskExecutor{
				Runner: cmdRunner,
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			},
			tasks.PkgInstalled:   pkgTaskExecutor,
			tasks.PkgRemoved:     pkgTaskExecutor,
			tasks.PkgUpgraded:    pkgTaskExecutor,
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
