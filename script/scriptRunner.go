package script

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/cloudradar-monitoring/tacoscript/exec"
	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type Runner struct {
	ExecutorRouter tasks.ExecutorRouter
	DataProvider   FileDataProvider
}

func (r Runner) Run(ctx context.Context, scripts tasks.Scripts, globalAbortOnError bool, output io.Writer) error {
	SortScriptsRespectingRequirements(scripts)

	result := Result{}
	scriptStart := time.Now()

	summary := scriptSummary{}

	for _, script := range scripts {
		summary.Total += len(script.Tasks)
	}

	for _, script := range scripts {
		logrus.Debugf("will run script '%s'", script.ID)
		abort := false
		for _, task := range script.Tasks {
			taskStart := time.Now()
			executor, err := r.ExecutorRouter.GetExecutor(task)
			if err != nil {
				return err
			}

			logrus.Debugf("will run task '%s' at path '%s'", task.GetTypeName(), task.GetPath())

			res := executor.Execute(ctx, task)

			logrus.Debugf("finished task '%s' at path '%s', result: %s", task.GetTypeName(), task.GetPath(), res.String())

			if res.Succeeded() {
				summary.Succeeded++
			} else {
				summary.Failed++
			}

			summary.TotalTasksRun++

			name := ""
			comment := ""
			changeMap := make(map[string]interface{})

			if cmdRunTask, ok := task.(*tasks.CmdRunTask); ok {
				// summary and changeMap will be updated
				name, comment, abort = handleCmdRunResults(cmdRunTask, &summary, &res, changeMap)
			}

			if pkgTask, ok := task.(*tasks.PkgTask); ok {
				name = pkgTask.Named.Name
				comment = res.Comment
				if res.Err == nil && !pkgTask.Updated {
					comment = "Package not updated " + res.SkipReason
				}
			}

			if winRegTask, ok := task.(*tasks.WinRegTask); ok {
				name = winRegTask.RegPath + `\` + winRegTask.Name
				comment = res.Comment
				if res.Err == nil && !winRegTask.Updated {
					comment = "Windows registry not updated " + res.SkipReason
				}
			}

			if managedTask, ok := task.(*tasks.FileManagedTask); ok {
				name = managedTask.Name
				comment = res.Comment
				if res.Err == nil && !managedTask.Updated {
					comment = "File not changed " + res.SkipReason
				}
			}

			if replaceTask, ok := task.(*tasks.FileReplaceTask); ok {
				name = replaceTask.Name
				comment = res.Comment
				if res.Err == nil && !replaceTask.Updated {
					comment = "File not changed " + res.SkipReason
				}
			}

			if realVNCServerTask, ok := task.(*tasks.RealVNCServerTask); ok {
				comment = res.Comment
				if res.Err == nil && !realVNCServerTask.Updated {
					comment = "Config not changed " + res.SkipReason
				}
			}

			if len(res.Changes) > 0 {
				for k, v := range res.Changes {
					changeMap[k] = v
				}
				summary.Changes++
			}

			errString := ""
			if res.Err != nil {
				errString = res.Err.Error()
			}

			result.Results = append(result.Results, taskResult{
				ID:       script.ID,
				Function: task.GetTypeName(),
				Name:     name,
				Result:   res.Succeeded(),
				Comment:  comment,
				Started:  onlyTime(taskStart),
				Duration: res.Duration,
				Changes:  changeMap,
				Error:    errString,
			})
		}

		if abort || globalAbortOnError {
			logrus.Debug("aborting due to task failure")
			summary.Aborted = summary.Total - summary.TotalTasksRun
			break
		}
		logrus.Debugf("finished script '%s'", script.ID)
	}

	summary.Script = r.DataProvider.Path
	summary.TotalRunTime = time.Since(scriptStart)
	result.Summary = summary

	y, err := yaml.Marshal(result)
	if err != nil {
		return err
	}

	fmt.Fprintln(output, string(y))

	if summary.Aborted > 0 || summary.Failed > 0 {
		return fmt.Errorf("%d aborted, %d failed", summary.Aborted, summary.Failed)
	}

	return nil
}

func handleCmdRunResults(
	cmdRunTask *tasks.CmdRunTask,
	summary *scriptSummary,
	res *tasks.ExecutionResult,
	changeMap map[string]interface{}) (name string, comment string, abort bool) {
	name = strings.Join(cmdRunTask.Named.GetNames(), "; ")

	if !res.IsSkipped {
		comment = `Command "` + name + `" run`
		changeMap["pid"] = res.Pid
		if runErr, ok := res.Err.(exec.RunError); ok {
			changeMap["retcode"] = runErr.ExitCode
		}

		changeMap["stderr"] = strings.TrimSpace(strings.ReplaceAll(res.StdErr, "\r\n", "\n"))
		changeMap["stdout"] = strings.TrimSpace(strings.ReplaceAll(res.StdOut, "\r\n", "\n"))

		if exec.IsPowerShell(cmdRunTask.Shell) {
			changeMap["stdout"] = powershellUnquote(changeMap["stdout"].(string))
		}
		summary.Changes++
	} else {
		comment = `Command skipped: ` + res.SkipReason
	}

	if cmdRunTask.AbortOnError && !res.Succeeded() {
		abort = true
		summary.Aborted = summary.Total - summary.TotalTasksRun
	}

	return name, comment, abort
}

// stdout from multiline powershell scripts often includes trailing spaces on each line.
// when encoded as yaml, the result does not look pretty. this function strips trailing whitespace
func powershellUnquote(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}
