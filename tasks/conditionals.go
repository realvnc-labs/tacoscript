package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
)

func CheckConditionals(ctx *tacoexec.Context, fsManager FsManager, runner tacoexec.Runner, task Task) (
	skipExecutionReason string, err error) {
	defer func() {
		if skipExecutionReason != "" {
			logrus.Debugf(skipExecutionReason+", will skip the execution of %s", task.GetPath())
		}
	}()

	isExists, filename, err := checkMissingFileCondition(fsManager, task)
	if err != nil {
		return "", err
	}

	if isExists {
		skipExecutionReason = fmt.Sprintf("file %s exists", filename)
		return skipExecutionReason, nil
	}

	isSuccess, err := checkOnlyIfs(ctx, runner, task)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		skipExecutionReason = onlyIfConditionFailedReason
		return skipExecutionReason, nil
	}

	isExpectationSuccess, err := checkUnless(ctx, runner, task)
	if err != nil {
		return "", err
	}

	if !isExpectationSuccess {
		skipExecutionReason = "unless condition is true"
		return skipExecutionReason, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", task.GetPath())
	return "", nil
}

func checkMissingFileCondition(fsManager FsManager, task Task) (isExists bool, filename string, err error) {
	createsFilesList := task.GetCreatesFilesList()

	if len(createsFilesList) == 0 {
		return false, "", nil
	}

	for _, missingFileCondition := range createsFilesList {
		if missingFileCondition == "" {
			continue
		}
		isExists, err = fsManager.FileExists(missingFileCondition)
		if err != nil {
			err = fmt.Errorf("failed to check if file '%s' exists: %w", missingFileCondition, err)
			return false, "", err
		}

		if isExists {
			logrus.Debugf("file '%s' exists", missingFileCondition)
			return true, missingFileCondition, nil
		}
		logrus.Debugf("file '%s' doesn't exist", missingFileCondition)
	}

	return false, "", nil
}

func runCommands(ctx *tacoexec.Context, runner tacoexec.Runner, cmds []string) (err error) {
	newCtx := ctx.Copy()
	newCtx.Cmds = cmds
	return runner.Run(&newCtx)
}

func checkUnless(ctx *tacoexec.Context, runner tacoexec.Runner, task Task) (isExpectationSuccess bool, err error) {
	cmds := task.GetUnlessCmds()
	if len(cmds) == 0 {
		return true, nil
	}

	err = runCommands(ctx, runner, cmds)
	if err != nil {
		runErr, isRunErr := err.(tacoexec.RunError)
		if isRunErr {
			logrus.Infof("will continue %s since at least one unless condition has failed: %v", task.GetPath(), runErr)
			return true, nil
		}

		return false, err
	}

	logrus.Debugf("any unless condition didn't fail for task '%s'", task.GetPath())
	return false, nil
}

func checkOnlyIfs(ctx *tacoexec.Context, runner tacoexec.Runner, task Task) (isSuccess bool, err error) {
	cmds := task.GetOnlyIfCmds()
	if len(cmds) == 0 {
		return true, nil
	}

	err = runCommands(ctx, runner, cmds)
	if err != nil {
		runErr, isRunErr := err.(tacoexec.RunError)
		if isRunErr {
			logrus.Debugf("will skip %s since onlyif condition has failed: %v", task.GetPath(), runErr)
			return false, nil
		}

		return false, err
	}

	return true, nil
}
