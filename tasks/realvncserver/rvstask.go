package realvncserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"

	tacoexec "github.com/realvnc-labs/tacoscript/exec"
	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/shared/executionresult"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
)

const (
	TaskTypeConfigUpdate = "realvnc_server.config_update"

	ServiceServerMode = "Service"
	UserServerMode    = "User"
	VirtualServerMode = "Virtual"
	// TODO: (rs): this is hack to get around no admin user when running github actions on windows.
	// it can be removed when we re-introduce User server mode as the e2e tests can use that instead
	// of Service server mode. will only work with windows.
	TestServerMode = "Test"
)

type Task struct {
	TypeName string // TaskType
	Path     string // TaskName

	// the Task uses the field status tracking capability (the second field of the tag) to indicate
	// whether a field is a realvnc server parameter or not. if set to true then it is a config parameter
	// and it will be used to update the realvnc server config. if not tracked, then it means the field
	// is used for the task itself and is NOT a realvnc config param.
	Encryption          string `taco:"encryption,true"`
	Authentication      string `taco:"authentication,true"`
	Permissions         string `taco:"permissions,true"` // multiple key:value pairs delimited by commas
	QueryConnect        bool   `taco:"query_connect,true"`
	QueryOnlyIfLoggedOn bool   `taco:"query_only_if_logged_on,true"`
	QueryConnectTimeout int    `taco:"query_connect_timeout,true"` // seconds
	BlankScreen         bool   `taco:"blank_screen,true"`
	ConnNotifyTimeout   int    `taco:"conn_notify_timeout,true"` // seconds
	ConnNotifyAlways    bool   `taco:"conn_notify_always,true"`
	IdleTimeout         int    `taco:"idle_timeout,true"` // seconds
	Log                 string `taco:"log,true"`
	CaptureMethod       int    `taco:"capture_method,true"`

	ConfigFile          string `taco:"config_file"` // config file path for non-windows
	ServerMode          string `taco:"server_mode"` // server mode for windows (registry keys)
	ReloadExecPath      string `taco:"reload_exec_path"`
	SkipReload          bool   `taco:"skip_reload"`
	UseVNCLicenseReload bool   `taco:"use_vnclicense_reload"`

	Backup     string `taco:"backup"`
	SkipBackup bool   `taco:"skip_backup"`

	Require []string `taco:"require"`

	Creates []string `taco:"creates"`
	OnlyIf  []string `taco:"onlyif"`
	Unless  []string `taco:"unless"`

	Shell string `taco:"shell"`

	fieldMapper  fieldstatus.NameMapper
	fieldTracker fieldstatus.Tracker

	// was the config updated?
	Updated bool
}

var (
	ErrFieldNotFound         = errors.New("task field not found")
	ErrFieldTypeNotSupported = errors.New("task field type not supported")

	// make sure we support the field tracker interface
	_ tasks.TaskWithFieldTracker = new(Task)
)

func (t *Task) SetMapper(mapper fieldstatus.NameMapper) {
	t.fieldMapper = mapper
}

func (t *Task) SetTracker(tracker fieldstatus.Tracker) {
	t.fieldTracker = tracker
}

func (t *Task) GetTypeName() string {
	return t.TypeName
}

func (t *Task) GetRequirements() []string {
	return t.Require
}

func (t *Task) GetPath() string {
	return t.Path
}

func (t *Task) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.TypeName, t.GetPath())
}

func (t *Task) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *Task) GetUnlessCmds() []string {
	return t.Unless
}

func (t *Task) GetCreatesFilesList() []string {
	return t.Creates
}

func (t *Task) getFieldValueAsString(fieldName string) (val string, err error) {
	rTaskValue := reflect.ValueOf(*t)

	// get field from the task
	field := rTaskValue.FieldByName(fieldName)

	// if empty field then we didn't find the field matching the name
	if field == (reflect.Value{}) {
		return "", ErrFieldNotFound
	}

	// get the field value based on type and return as string
	var valStr string
	switch field.Kind() { //nolint:exhaustive // default handler
	case reflect.Bool:
		valBool := field.Bool()
		valStr = fmt.Sprintf("%t", valBool)
	case reflect.String:
		valStr = field.String()
	case reflect.Int:
		valInt := field.Int()
		valStr = fmt.Sprintf("%d", valInt)
	default:
		return "", ErrFieldTypeNotSupported
	}

	// return the field value in string form
	return valStr, nil
}

type RvsConfigReloader interface {
	Reload(rvst *Task) (err error)
}

type RvstExecutor struct {
	FsManager tasks.FsManager
	Runner    tacoexec.Runner

	Reloader RvsConfigReloader
}

func (rvste *RvstExecutor) Execute(ctx context.Context, task tasks.CoreTask) executionresult.ExecutionResult {
	start := time.Now()

	logrus.Debugf("will trigger '%s' task", task.GetPath())
	execRes := executionresult.ExecutionResult{
		Changes: make(map[string]string),
	}

	rvst, ok := task.(*Task)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to Task", task)
		return execRes
	}

	execRes.Comment = "Config not changed"

	var stdoutBuf, stderrBuf bytes.Buffer
	execCtx := &tacoexec.Context{
		Ctx:          ctx,
		Path:         rvst.Path,
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		Shell:        rvst.Shell,
	}

	logrus.Debugf("will check if the task '%s' should be executed", task.GetPath())
	skipReason, err := tasks.CheckConditionals(execCtx, rvste.FsManager, rvste.Runner, rvst)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if skipReason != "" {
		logrus.Debugf("the task '%s' will be be skipped", task.GetPath())
		execRes.IsSkipped = true
		execRes.SkipReason = skipReason
		return execRes
	}

	addedCount, updatedCount, err := rvste.applyConfigChanges(rvst)
	if err != nil {
		execRes.Err = err
		return execRes
	}

	if addedCount > 0 || updatedCount > 0 {
		rvst.Updated = true
		execRes.Comment = "Config updated"
		execRes.Changes["count"] = fmt.Sprintf("%d config value change(s) applied", addedCount+updatedCount)

		if !rvst.SkipReload {
			if rvste.Reloader == nil {
				// use the task based config reload
				err = rvste.ReloadConfig(rvst)
			} else {
				// use custom reload
				err = rvste.Reloader.Reload(rvst)
			}
			if err != nil {
				execRes.Err = err
			}
		}
	}

	execRes.Duration = time.Since(start)

	logrus.Debugf("the task '%s' is finished for %v", task.GetPath(), execRes.Duration)
	return execRes
}
