package tasks

import "github.com/realvnc-labs/tacoscript/tasks/fieldstatus"

type Scripts []Script

type Script struct {
	ID    string
	Tasks []CoreTask
}

type CoreTask interface {
	GetTypeName() string
	Validate(goos string) error
	GetPath() string
	GetRequirements() []string
	GetCreatesFilesList() []string
	GetOnlyIfCmds() []string
	GetUnlessCmds() []string
}

// TaskWithFieldTracker allows the task access to both field mapper and tracker info.
// New interfaces will be required if there's a requirement for allowing access to only one or
// the other.
type TaskWithFieldTracker interface {
	SetMapper(mapper fieldstatus.NameMapper)
	SetTracker(tracker fieldstatus.Tracker)
}
