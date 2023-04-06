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
	"github.com/realvnc-labs/tacoscript/tasks/executionresult"
)

const (
	TaskTypeRealVNCServer = "realvnc_server.config_update"

	ServiceServerMode = "Service"
	UserServerMode    = "User"
	VirtualServerMode = "Virtual"
	// TODO: (rs): this is hack to get around no admin user when running github actions on windows.
	// it can be removed when we re-introduce User server mode as the e2e tests can use that instead
	// of Service server mode. will only work with windows.
	TestServerMode = "Test"
)

var (
	// these fields don't change the realvnc server config. they are only used by the task.
	RvstNoChangeFields = []string{"ConfigFile", "ServerMode", "ReloadExecPath", "SkipReload", "UseVNCLicenseReload", "Backup", "SkipBackup"}
)

type RvsTask struct {
	TypeName string // TaskType
	Path     string // TaskName

	Encryption          string `taco:"encryption"`
	Authentication      string `taco:"authentication"`
	Permissions         string `taco:"permissions"` // multiple key:value pairs delimited by commas
	QueryConnect        bool   `taco:"query_connect"`
	QueryOnlyIfLoggedOn bool   `taco:"query_only_if_logged_on"`
	QueryConnectTimeout int    `taco:"query_connect_timeout"` // seconds
	BlankScreen         bool   `taco:"blank_screen"`
	ConnNotifyTimeout   int    `taco:"conn_notify_timeout"` // seconds
	ConnNotifyAlways    bool   `taco:"conn_notify_always"`
	IdleTimeout         int    `taco:"idle_timeout"` // seconds
	Log                 string `taco:"log"`
	CaptureMethod       int    `taco:"capture_method"`

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

	Mapper  tasks.FieldNameMapper
	Tracker tasks.FieldStatusTracker

	// was replace file updated?
	Updated bool
}

var (
	ErrFieldNotFound         = errors.New("task field not found")
	ErrFieldTypeNotSupported = errors.New("task field type not supported")

	// make sure we support the field tracker interface
	_ tasks.TaskWithFieldTracker = new(RvsTask)
)

func (t *RvsTask) SetMapper(mapper tasks.FieldNameMapper) {
	t.Mapper = mapper
}

func (t *RvsTask) SetTracker(tracker tasks.FieldStatusTracker) {
	t.Tracker = tracker
}

func (t *RvsTask) IsChangeField(fieldName string) (excluded bool) {
	for _, noChangeField := range RvstNoChangeFields {
		if fieldName == noChangeField {
			return false
		}
	}
	return true
}

func (t *RvsTask) GetTypeName() string {
	return t.TypeName
}

func (t *RvsTask) GetRequirements() []string {
	return t.Require
}

func (t *RvsTask) GetPath() string {
	return t.Path
}

func (t *RvsTask) String() string {
	return fmt.Sprintf("task '%s' at path '%s'", t.TypeName, t.GetPath())
}

func (t *RvsTask) GetOnlyIfCmds() []string {
	return t.OnlyIf
}

func (t *RvsTask) GetUnlessCmds() []string {
	return t.Unless
}

func (t *RvsTask) GetCreatesFilesList() []string {
	return t.Creates
}

func (t *RvsTask) getFieldValueAsString(fieldName string) (val string, err error) {
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
	Reload(rvst *RvsTask) (err error)
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

	rvst, ok := task.(*RvsTask)
	if !ok {
		execRes.Err = fmt.Errorf("cannot convert task '%v' to RvsTask", task)
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
