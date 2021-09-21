package script

import (
	"context"
	"fmt"
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

	for _, script := range scripts {
		logrus.Debugf("will run script '%s'", script.ID)
		abort := false
		for _, task := range script.Tasks {
			taskStart := time.Now()
			executr, err := r.ExecutorRouter.GetExecutor(task)
			if err != nil {
				return err
			}

			logrus.Debugf("will run task '%s' at path '%s'", task.GetName(), task.GetPath())
			res := executr.Execute(ctx, task)

			logrus.Debugf("finished task '%s' at path '%s', result: %s", task.GetName(), task.GetPath(), res.String())

			if res.Succeeded() {
				succeeded++
			} else {
				failed++
			}

			tasksRun++

			changeMap := make(map[string]string)

			if !res.IsSkipped {
				changeMap["pid"] = intsToString(res.Pids)
				if runErr, ok := res.Err.(exec.RunError); ok {
					changeMap["retcode"] = fmt.Sprintf("%d", runErr.ExitCode)
				}
			}

			if len(res.Changes) > 0 {
				for k, v := range res.Changes {
					changeMap[k] = v
				}
			}

			result.Results = append(result.Results, taskResult{
				ID:       script.ID,
				Function: task.GetName(),
				Name:     res.Name,
				Result:   res.Succeeded(),
				Comment:  res.Comment,
				Started:  onlyTime(taskStart),
				Duration: res.Duration,
				Changes:  changeMap,
			})
		}

		if abort || globalAbortOnError {
			logrus.Debugf("aborting due to task failure")
			break
		}
		logrus.Debugf("finished script '%s'", script.ID)
	}

	result.Summary = scriptSummary{
		Config:            r.DataProvider.Path,
		Succeeded:         succeeded,
		Failed:            failed,
		Changes:           changes,
		TotalFunctionsRun: tasksRun,
		TotalRunTime:      time.Since(scriptStart),
	}

	y, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(y))

	return nil
}
