package script

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type Runner struct {
}

func (r Runner) Run(ctx context.Context, scripts tasks.Scripts) error {
	SortScriptsRespectingRequirements(scripts)

	for _, script := range scripts {
		logrus.Infof("will run script '%s'", script.ID)
		for _, task := range script.Tasks {
			logrus.Debugf("will run task '%s' at path '%s'", task.GetName(), task.GetPath())
			res := task.Execute(ctx)
			logrus.Infof("finished task '%s' at path '%s', result: %s", task.GetName(), task.GetPath(), res)
			if res.Err != nil {
				return res.Err
			}
		}
		logrus.Infof("finished script '%s'", script.ID)
	}

	return nil
}
