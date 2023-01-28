package script

import (
	"context"
	"io"

	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/pkgmanager"
	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

// RunScript main entry point for the script execution
func RunScript(scriptPath string, abortOnError bool, output io.Writer) error {
	fileDataProvider := FileDataProvider{
		Path: scriptPath,
	}

	parser := Builder{
		DataProvider: fileDataProvider,
		TaskBuilder: tasks.NewBuilderRouter(map[string]tasks.Builder{
			tasks.TaskTypeCmdRun:  &tasks.CmdRunTaskBuilder{},
			tasks.FileManaged:     &tasks.FileManagedTaskBuilder{},
			tasks.FileReplace:     &tasks.FileReplaceTaskBuilder{},
			tasks.PkgInstalled:    &tasks.PkgTaskBuilder{},
			tasks.PkgRemoved:      &tasks.PkgTaskBuilder{},
			tasks.PkgUpgraded:     &tasks.PkgTaskBuilder{},
			tasks.WinRegPresent:   &tasks.WinRegTaskBuilder{},
			tasks.WinRegAbsent:    &tasks.WinRegTaskBuilder{},
			tasks.WinRegAbsentKey: &tasks.WinRegTaskBuilder{},
		}),
		TemplateVariablesProvider: utils.OSDataProvider{},
	}

	cmdRunner := exec.SystemRunner{
		SystemAPI: exec.OSApi{},
	}

	pkgTaskManager := pkgmanager.PackageTaskManager{
		Runner:                          cmdRunner,
		ManagementCmdsProviderBuildFunc: pkgmanager.BuildManagementCmdsProviders,
	}

	pkgTaskExecutor := &tasks.PkgTaskExecutor{
		PackageManager: pkgTaskManager,
		Runner:         cmdRunner,
		FsManager:      &utils.FsManager{},
	}

	winRegTaskExecutor := &tasks.WinRegTaskExecutor{
		Runner:    cmdRunner,
		FsManager: &utils.FsManager{},
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
			tasks.FileReplace: &tasks.FileReplaceTaskExecutor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			tasks.PkgInstalled:    pkgTaskExecutor,
			tasks.PkgRemoved:      pkgTaskExecutor,
			tasks.PkgUpgraded:     pkgTaskExecutor,
			tasks.WinRegPresent:   winRegTaskExecutor,
			tasks.WinRegAbsent:    winRegTaskExecutor,
			tasks.WinRegAbsentKey: winRegTaskExecutor,
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

	err = runner.Run(context.Background(), scripts, abortOnError, output)
	return err
}
