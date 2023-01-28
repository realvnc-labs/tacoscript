package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/cloudradar-monitoring/tacoscript/exec"
)

func shouldCheckConditionals(ctx *tacoexec.Context, fsManager FsManager, runner tacoexec.Runner, task Task) (
	skipExecutionReason string, err error) {
	isExists, filename, err := checkMissingFileCondition(fsManager, task)
	if err != nil {
		return "", err
	}

	if isExists {
		skipExecutionReason = fmt.Sprintf("file %s exists", filename)
		logrus.Debugf(skipExecutionReason+", will skip the execution of %s", task.GetPath())
		return skipExecutionReason, nil
	}

	isSuccess, err := checkOnlyIfs(ctx, runner, task)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return onlyIfConditionFailedReason, nil
	}

	isExpectationSuccess, err := checkUnless(ctx, runner, task)
	if err != nil {
		return "", err
	}

	if !isExpectationSuccess {
		skipExecutionReason = "unless condition is true"
		logrus.Debugf(skipExecutionReason+", will skip %s", task.GetPath())
		return skipExecutionReason, nil
	}

	logrus.Debugf("all execution conditions are met, will continue %s", task.GetPath())
	return "", nil
}

func checkMissingFileCondition(fsManager FsManager, task Task) (isExists bool, filename string, err error) {
	createsFilesList := task.GetCreatesFilesList()

	if len(createsFilesList) == 0 {
		return
	}

	for _, missingFileCondition := range createsFilesList {
		if missingFileCondition == "" {
			continue
		}
		isExists, err = fsManager.FileExists(missingFileCondition)
		if err != nil {
			err = fmt.Errorf("failed to check if file '%s' exists: %w", missingFileCondition, err)
			return
		}

		if isExists {
			logrus.Debugf("file '%s' exists", missingFileCondition)
			return true, missingFileCondition, nil
		}
		logrus.Debugf("file '%s' doesn't exist", missingFileCondition)
	}

	return
}

func checkUnless(ctx *tacoexec.Context, runner tacoexec.Runner, task Task) (isExpectationSuccess bool, err error) {
	unlessCmds := task.GetUnlessCmds()

	if len(unlessCmds) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = unlessCmds

	err = runner.Run(&newCtx)

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
	onlyIfCmds := task.GetOnlyIfCmds()

	if len(onlyIfCmds) == 0 {
		return true, nil
	}

	newCtx := ctx.Copy()

	newCtx.Cmds = onlyIfCmds

	err = runner.Run(&newCtx)

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
