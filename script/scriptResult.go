package script

import (
	"time"
)

type scriptResult struct {
	Results []taskResult

	Summary scriptSummary
}

type taskResult struct {
	ID       string `yaml:"ID"`
	Function string `yaml:"Function"`
	Name     string `yaml:"Name"`
	Result   bool   `yaml:"Result"`
	Comment  string `yaml:"Comment,omitempty"`
	Error    string `yaml:"Error,omitempty"`

	Started  onlyTime      `yaml:"Started"`
	Duration time.Duration `yaml:"Duration"`

	Changes map[string]string `yaml:"Changes,omitempty"` // map for custom key-val data depending on type
}

type scriptSummary struct {
	Config            string        `yaml:"Config"`
	Succeeded         int           `yaml:"Succeeded"`
	Failed            int           `yaml:"Failed"`
	Aborted           int           `yaml:"Aborted"`
	Changes           int           `yaml:"Changes"`
	TotalFunctionsRun int           `yaml:"TotalFunctionsRun"`
	TotalRunTime      time.Duration `yaml:"TotalRunTime"`
}

const stampMicro = "15:04:05.000000"

type onlyTime time.Time

func (c onlyTime) MarshalYAML() (interface{}, error) {
	return time.Time(c).Format(stampMicro), nil
}
