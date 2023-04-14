package script

import (
	"context"
	"io"

	"github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun/crtbuilder"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged/fmtbuilder"
	"github.com/realvnc-labs/tacoscript/tasks/filereplace"
	"github.com/realvnc-labs/tacoscript/tasks/filereplace/frtbuilder"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask/pkgbuilder"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver/rvstbuilder"
	"github.com/realvnc-labs/tacoscript/tasks/shared/builder"
	"github.com/realvnc-labs/tacoscript/tasks/support/pkgmanager"
	"github.com/realvnc-labs/tacoscript/tasks/winreg"
	"github.com/realvnc-labs/tacoscript/tasks/winreg/wrtbuilder"
	"github.com/realvnc-labs/tacoscript/utils"

	"github.com/realvnc-labs/tacoscript/tasks"
)

// RunScript main entry point for the script execution
func RunScript(scriptPath string, abortOnError bool, output io.Writer) error {
	fileDataProvider := FileDataProvider{
		Path: scriptPath,
	}

	parser := Builder{
		DataProvider: fileDataProvider,
		TaskBuilder: builder.NewBuilderRouter(map[string]builder.Builder{
			cmdrun.TaskType:                    &crtbuilder.TaskBuilder{},
			filemanaged.TaskType:               &fmtbuilder.TaskBuilder{},
			filereplace.TaskType:               &frtbuilder.TaskBuilder{},
			realvncserver.TaskTypeConfigUpdate: &rvstbuilder.TaskBuilder{},
			pkgtask.TaskTypePkgInstalled:       &pkgbuilder.TaskBuilder{},
			pkgtask.TaskTypePkgRemoved:         &pkgbuilder.TaskBuilder{},
			pkgtask.TaskTypePkgUpgraded:        &pkgbuilder.TaskBuilder{},
			winreg.TaskTypeWinRegPresent:       &wrtbuilder.TaskBuilder{},
			winreg.TaskTypeWinRegAbsent:        &wrtbuilder.TaskBuilder{},
			winreg.TaskTypeWinRegAbsentKey:     &wrtbuilder.TaskBuilder{},
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

	pkgTaskExecutor := &pkgtask.Executor{
		PackageManager: pkgTaskManager,
		Runner:         cmdRunner,
		FsManager:      &utils.FsManager{},
	}

	winRegTaskExecutor := &winreg.Executor{
		Runner:    cmdRunner,
		FsManager: &utils.FsManager{},
	}

	execRouter := tasks.ExecutorRouter{
		Executors: map[string]tasks.Executor{
			cmdrun.TaskType: &cmdrun.Executor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			filemanaged.TaskType: &filemanaged.Executor{
				Runner:      cmdRunner,
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			},
			filereplace.TaskType: &filereplace.Executor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			realvncserver.TaskTypeConfigUpdate: &realvncserver.Executor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			pkgtask.TaskTypePkgInstalled:   pkgTaskExecutor,
			pkgtask.TaskTypePkgRemoved:     pkgTaskExecutor,
			pkgtask.TaskTypePkgUpgraded:    pkgTaskExecutor,
			winreg.TaskTypeWinRegPresent:   winRegTaskExecutor,
			winreg.TaskTypeWinRegAbsent:    winRegTaskExecutor,
			winreg.TaskTypeWinRegAbsentKey: winRegTaskExecutor,
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
