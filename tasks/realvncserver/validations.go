package realvncserver

import (
	"errors"
	"os/user"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/utils"
)

const (
	DefaultLinuxServiceServerModeConfigFile  = `/root/.vnc/config.d/vncserver-x11`
	DefaultDarwinServiceServerModeConfigFile = `/var/root/.vnc/config.d/vncserver`
	// the user's home dir will be prepended to the values below
	DefaultUserServerModeConfigFile    = `/.vnc/config.d/vncserver-x11`
	DefaultVirtualServerModeConfigFile = `/.vnc/config.d/vncserver-x11-virtual`
)

var (
	AllowableEncryptionValues        = []any{"AlwaysOn", "PreferOn", "AlwaysMaximum", "PreferOff", "AlwaysOff"}
	AllowableAuthenticationValues    = []any{"VncAuth", "SystemAuth", "InteractiveSystemAuth", "SingleSignOn", "Certificate", "Radius", "None"}
	AllowableFeaturePermissionsChars = "!-svkpctrhwdqf"
	AllowableLogTargets              = []any{"stderr", "file", "EventLog", "syslog"}
	AllowableServerModes             = []any{ServiceServerMode, UserServerMode, VirtualServerMode, TestServerMode}
	AllowableLogLevels               = []any{0, 10, 30, 100}

	ErrInvalidNameFieldMsg          = "invalid task name"
	ErrConfigFileMustBeSpecifiedMsg = "the config_file param must be specified when updating a realvnc server config on linux or mac"

	ErrInvalidEncryptionValueMsg                = "invalid encryption value"
	ErrInvalidAuthenticationValueMsg            = "invalid authentication value"
	ErrEmptyAuthenticationValueMsg              = "authentication value cannot be empty"
	ErrInvalidPermisssionsMsg                   = "invalid permissions value"
	ErrInvalidLogsValueMsg                      = "invalid log value"
	ErrInvalidCaptureMethodValueMsg             = "invalid captureMethod value"
	ErrUnknownServerModeMsg                     = "unknown server mode"
	ErrServerModeCannotBeVirtualWhenNotLinuxMsg = "server mode cannot be virtual when not running Linux"
)

func (t *Task) Validate(goos string) error {
	errs := &utils.Errors{}

	err := tasks.ValidateRequired(t.Path, t.Path+"."+tasks.NameField)
	errs.Add(err)

	err = t.ValidateServerModeField(goos)
	if err != nil {
		errs.Add(err)
	}

	err = t.ValidateConfigFileField(goos)
	if err != nil {
		errs.Add(err)
		return errs.ToError()
	}

	if t.shouldValidate(tasks.EncryptionField) {
		err = t.ValidateEncryptionField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(tasks.AuthenticationField) {
		err = t.ValidateAuthenticationField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(tasks.PermissionsField) {
		err = t.ValidatePermissionsField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(tasks.LogField) {
		err = t.ValidateLogField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(tasks.CaptureMethodField) {
		err = t.ValidateCaptureMethodField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.Backup == "" {
		t.Backup = "bak"
	}

	return errs.ToError()
}

func (t *Task) shouldValidate(fieldKey string) (should bool) {
	fieldName := t.fieldMapper.GetFieldName(fieldKey)
	return t.fieldTracker.HasNewValue(fieldName) && !t.fieldTracker.ShouldClear(fieldName)
}

func (t *Task) ValidateConfigFileField(goos string) error {
	if goos != "windows" && t.ConfigFile == "" {
		switch t.ServerMode {
		case ServiceServerMode:
			if goos == "darwin" {
				t.ConfigFile = DefaultDarwinServiceServerModeConfigFile
			} else {
				t.ConfigFile = DefaultLinuxServiceServerModeConfigFile
			}
		case UserServerMode:
			homeDir, err := getUserHomeDir()
			if err != nil {
				return err
			}
			t.ConfigFile = homeDir + DefaultUserServerModeConfigFile
		case VirtualServerMode:
			homeDir, err := getUserHomeDir()
			if err != nil {
				return err
			}
			t.ConfigFile = homeDir + DefaultVirtualServerModeConfigFile
		}
		err := validation.Validate(t.ConfigFile,
			validation.Required.Error(ErrConfigFileMustBeSpecifiedMsg),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func getUserHomeDir() (homeDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDir = usr.HomeDir
	return homeDir, nil
}

func (t *Task) ValidateServerModeField(goos string) error {
	if t.ServerMode == "" {
		t.ServerMode = ServiceServerMode
		return nil
	}

	caser := cases.Title(language.AmericanEnglish)
	serverMode := caser.String(t.ServerMode)

	// TODO: (rs): Remove this check when user and virtual server modes reintroduced
	if serverMode != ServiceServerMode && serverMode != TestServerMode {
		return errors.New("user and virtual server modes will be supported in a future release")
	}

	err := validation.Validate(serverMode,
		validation.In(AllowableServerModes...).Error(ErrUnknownServerModeMsg),
	)
	if err != nil {
		return err
	}

	if serverMode == VirtualServerMode {
		if goos != "linux" {
			return errors.New(ErrServerModeCannotBeVirtualWhenNotLinuxMsg)
		}
		t.UseVNCLicenseReload = true
	}

	t.ServerMode = serverMode
	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Encryption
func (t *Task) ValidateEncryptionField() error {
	err := validation.Validate(t.Encryption,
		validation.In(AllowableEncryptionValues...).Error(ErrInvalidEncryptionValueMsg),
	)
	if err != nil {
		return err
	}
	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Authentication
func (t *Task) ValidateAuthenticationField() error {
	authValue := t.Authentication
	authParts := strings.Split(authValue, ",")
	for _, authPart := range authParts {
		authTypeParts := strings.Split(authPart, "+")
		for _, authTypePart := range authTypeParts {
			err := validation.Validate(strings.TrimSpace(authTypePart),
				validation.Required.Error(ErrEmptyAuthenticationValueMsg).Error(ErrInvalidAuthenticationValueMsg),
				validation.In(AllowableAuthenticationValues...).Error(ErrInvalidAuthenticationValueMsg))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Permissions
func (t *Task) ValidatePermissionsField() error {
	permsValue := t.Permissions
	permsParts := strings.Split(permsValue, ",")
	for _, permPart := range permsParts {
		permParts := strings.Split(permPart, ":")
		if len(permParts) != 2 {
			return errors.New(ErrInvalidPermisssionsMsg)
		}

		features := strings.TrimSpace(permParts[1])

		for _, feature := range features {
			if !strings.Contains(AllowableFeaturePermissionsChars, string(feature)) {
				return errors.New(ErrInvalidPermisssionsMsg)
			}
		}
	}

	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Log
func (t *Task) ValidateLogField() error {
	logsValue := t.Log
	logsParts := strings.Split(logsValue, ",")
	for _, logPart := range logsParts {
		logParts := strings.Split(logPart, ":")
		if len(logParts) != 3 {
			return errors.New(ErrInvalidLogsValueMsg)
		}

		area := strings.TrimSpace(logParts[0])
		target := strings.TrimSpace(logParts[1])
		level := strings.TrimSpace(logParts[2])

		err := validation.Validate(area, validation.Required.Error(ErrInvalidLogsValueMsg))
		if err != nil {
			return err
		}

		err = validation.Validate(target,
			validation.Required.Error(ErrInvalidLogsValueMsg),
			validation.In(AllowableLogTargets...).Error(ErrInvalidLogsValueMsg),
		)
		if err != nil {
			return err
		}

		levelInt, err := strconv.Atoi(level)
		if err != nil {
			return errors.New(ErrInvalidLogsValueMsg)
		}

		err = validation.Validate(levelInt,
			validation.Required.Error(ErrInvalidLogsValueMsg),
			validation.In(AllowableLogLevels...).Error(ErrInvalidLogsValueMsg),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Capture_Method
func (t *Task) ValidateCaptureMethodField() error {
	method := t.CaptureMethod

	err := validation.Validate(method,
		validation.Min(0).Error(ErrInvalidCaptureMethodValueMsg),
		validation.Max(2).Error(ErrInvalidCaptureMethodValueMsg),
	)
	if err != nil {
		return err
	}
	return nil
}
