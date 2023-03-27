package script

import (
	"context"
	"io"

	"github.com/realvnc-labs/tacoscript/builder"
	"github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/pkgmanager"
	"github.com/realvnc-labs/tacoscript/tasks/cmdrun"
	cmdrunbuilder "github.com/realvnc-labs/tacoscript/tasks/cmdrun/builder"
	"github.com/realvnc-labs/tacoscript/tasks/filemanaged"
	filemanagedbuilder "github.com/realvnc-labs/tacoscript/tasks/filemanaged/builder"
	"github.com/realvnc-labs/tacoscript/tasks/filereplace"
	filereplacebuilder "github.com/realvnc-labs/tacoscript/tasks/filereplace/builder"
	"github.com/realvnc-labs/tacoscript/tasks/pkgtask"
	pkgtaskbuilder "github.com/realvnc-labs/tacoscript/tasks/pkgtask/builder"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	realvncserverbuilder "github.com/realvnc-labs/tacoscript/tasks/realvncserver/builder"
	"github.com/realvnc-labs/tacoscript/tasks/winreg"
	winregbuilder "github.com/realvnc-labs/tacoscript/tasks/winreg/builder"
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
			cmdrun.TaskTypeCmdRun:               &cmdrunbuilder.CmdRunTaskBuilder{},
			filemanaged.TaskTypeFileManaged:     &filemanagedbuilder.FileManagedTaskBuilder{},
			filereplace.TaskTypeFileReplace:     &filereplacebuilder.FileReplaceTaskBuilder{},
			realvncserver.TaskTypeRealVNCServer: &realvncserverbuilder.RealVNCServerTaskBuilder{},
			pkgtask.TaskTypePkgInstalled:        &pkgtaskbuilder.PkgTaskBuilder{},
			pkgtask.TaskTypePkgRemoved:          &pkgtaskbuilder.PkgTaskBuilder{},
			pkgtask.TaskTypePkgUpgraded:         &pkgtaskbuilder.PkgTaskBuilder{},
			winreg.TaskTypeWinRegPresent:        &winregbuilder.WinRegTaskBuilder{},
			winreg.TaskTypeWinRegAbsent:         &winregbuilder.WinRegTaskBuilder{},
			winreg.TaskTypeWinRegAbsentKey:      &winregbuilder.WinRegTaskBuilder{},
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

	pkgTaskExecutor := &pkgtask.PtExecutor{
		PackageManager: pkgTaskManager,
		Runner:         cmdRunner,
		FsManager:      &utils.FsManager{},
	}

	winRegTaskExecutor := &winreg.WrtExecutor{
		Runner:    cmdRunner,
		FsManager: &utils.FsManager{},
	}

	execRouter := tasks.ExecutorRouter{
		Executors: map[string]tasks.Executor{
			cmdrun.TaskTypeCmdRun: &cmdrun.CrtExecutor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			filemanaged.TaskTypeFileManaged: &filemanaged.FmtExecutor{
				Runner:      cmdRunner,
				FsManager:   &utils.FsManager{},
				HashManager: &utils.HashManager{},
			},
			filereplace.TaskTypeFileReplace: &filereplace.FrtExecutor{
				Runner:    cmdRunner,
				FsManager: &utils.FsManager{},
			},
			realvncserver.TaskTypeRealVNCServer: &realvncserver.RvstExecutor{
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
