package script

import (
	"context"
	"fmt"
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

func (r Runner) Run(ctx context.Context, scripts tasks.Scripts, globalAbortOnError bool) error {
	SortScriptsRespectingRequirements(scripts)

	result := scriptResult{}
	scriptStart := time.Now()

	succeeded := 0
	failed := 0
	tasksRun := 0
	changes := 0
	aborted := 0

	total := 0

	for _, script := range scripts {
		total += len(script.Tasks)
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

			logrus.Debugf("will run task '%s' at path '%s'", task.GetName(), task.GetPath())
			res := executor.Execute(ctx, task)

			logrus.Debugf("finished task '%s' at path '%s', result: %s", task.GetName(), task.GetPath(), res.String())

			if res.Succeeded() {
				succeeded++
			} else {
				failed++
			}

			tasksRun++

			name := ""
			comment := ""
			changeMap := make(map[string]string)

			if cmdRunTask, ok := task.(*tasks.CmdRunTask); ok {
				name = strings.Join(cmdRunTask.GetNames(), "; ")

				if !res.IsSkipped {
					comment = `Command "` + name + `" run`
					changeMap["pid"] = fmt.Sprintf("%d", res.Pid)
					if runErr, ok := res.Err.(exec.RunError); ok {
						changeMap["retcode"] = fmt.Sprintf("%d", runErr.ExitCode)
					}

					changeMap["stderr"] = strings.TrimSpace(strings.ReplaceAll(res.StdErr, "\r\n", "\n"))
					changeMap["stdout"] = strings.TrimSpace(strings.ReplaceAll(res.StdOut, "\r\n", "\n"))

					if exec.IsPowerShell(cmdRunTask.Shell) {
						changeMap["stdout"] = powershellUnquote(changeMap["stdout"])
					}
					changes++
				}

				if cmdRunTask.AbortOnError && !res.Succeeded() {
					abort = true
					aborted = total - tasksRun
				}
			}
			if pkgTask, ok := task.(*tasks.PkgTask); ok {
				name = pkgTask.NamedTask.Name
			}

			if len(res.Changes) > 0 {
				for k, v := range res.Changes {
					changeMap[k] = v
				}
			}

			errString := ""
			if res.Err != nil {
				errString = res.Err.Error()
			}

			if managedTask, ok := task.(*tasks.FileManagedTask); ok {
				name = managedTask.Name
				if errString == "" {
					if managedTask.Updated {
						comment = "File " + name + " updated"
					} else {
						comment = "All files in creates exists"
					}
				}
			}

			result.Results = append(result.Results, taskResult{
				ID:       script.ID,
				Function: task.GetName(),
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
			aborted = total - tasksRun
			break
		}
		logrus.Debugf("finished script '%s'", script.ID)
	}

	result.Summary = scriptSummary{
		Script:        r.DataProvider.Path,
		Succeeded:     succeeded,
		Failed:        failed,
		Aborted:       aborted,
		Changes:       changes,
		TotalTasksRun: tasksRun,
		TotalRunTime:  time.Since(scriptStart),
	}

	y, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(y))

	if aborted > 0 || failed > 0 {
		return fmt.Errorf("%d aborted, %d failed", aborted, failed)
	}

	return nil
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
