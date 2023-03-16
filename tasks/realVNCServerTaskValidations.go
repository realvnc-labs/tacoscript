package tasks

import (
	"errors"
	"runtime"
	"strconv"
	"strings"

	"github.com/cloudradar-monitoring/tacoscript/utils"
	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	AllowableEncryptionValues        = []any{"AlwaysOn", "PreferOn", "AlwaysMaximum", "PreferOff", "AlwaysOff"}
	AllowableAuthenticationValues    = []any{"VncAuth", "SystemAuth", "InteractiveSystemAuth", "SingleSignOn", "Certificate", "Radius", "None"}
	AllowableFeaturePermissionsChars = "!-svkpctrhwdqf"
	AllowableLogTargets              = []any{"stderr", "file", "EventLog", "syslog"}
	AllowableServerModes             = []any{ServiceServerMode, UserServerMode}
	AllowableLogLevels               = []any{0, 10, 30, 100}

	ErrInvalidNameFieldMsg          = "invalid task name"
	ErrConfigFileMustBeSpecifiedMsg = "the config_file param must be specified when updating a realvnc server config on linux or mac"

	ErrInvalidEncryptionValueMsg     = "invalid Encryption value. Please see the RealVNC documentation for allowable values"
	ErrInvalidAuthenticationValueMsg = "invalid Authentication value. Please see the RealVNC documentation for allowable values"
	ErrEmptyAuthenticationValueMsg   = "authentication value cannot be empty. Please see the RealVNC documentation for more information"
	ErrInvalidPermisssionsMsg        = "invalid Permissions value. Please see the RealVNC documentation for more information"
	ErrInvalidLogsValueMsg           = "invalid Log value. Please see the RealVNC documentation for more information"
	ErrInvalidCaptureMethodValueMsg  = "invalid CaptureMethod value. Please see the RealVNC documentation for more information"
	ErrUnknownServerModeMsg          = "unknown server mode"
)

func (t *RealVNCServerTask) Validate() error {
	errs := &utils.Errors{}

	err := ValidateRequired(t.Path, t.Path+"."+NameField)
	errs.Add(err)

	if runtime.GOOS != "windows" {
		err = t.ValidateConfigFileField()
		if err != nil {
			errs.Add(err)
			return errs.ToError()
		}
	}

	err = t.ValidateServerModeField()
	if err != nil {
		errs.Add(err)
	}

	if t.shouldValidate(EncryptionField) {
		err = t.ValidateEncryptionField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(AuthenticationField) {
		err = t.ValidateAuthenticationField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(PermissionsField) {
		err = t.ValidatePermissionsField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(LogField) {
		err = t.ValidateLogField()
		if err != nil {
			errs.Add(err)
		}
	}

	if t.shouldValidate(CaptureMethodField) {
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

func (t *RealVNCServerTask) shouldValidate(fieldKey string) (should bool) {
	if t.tracker == nil {
		return false
	}
	return t.tracker.HasNewValue(fieldKey) && !t.tracker.ShouldClear(fieldKey)
}

func (t *RealVNCServerTask) ValidateConfigFileField() error {
	if runtime.GOOS != "windows" {
		err := validation.Validate(t.ConfigFile,
			validation.Required.Error(ErrConfigFileMustBeSpecifiedMsg),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *RealVNCServerTask) ValidateServerModeField() error {
	if t.ServerMode == "" {
		t.ServerMode = ServiceServerMode
		return nil
	}

	caser := cases.Title(language.AmericanEnglish)
	serverMode := caser.String(t.ServerMode)
	err := validation.Validate(serverMode,
		validation.In(AllowableServerModes...).Error(ErrUnknownServerModeMsg),
	)
	t.ServerMode = serverMode
	if err != nil {
		return err
	}

	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Encryption
func (t *RealVNCServerTask) ValidateEncryptionField() error {
	err := validation.Validate(t.Encryption,
		validation.In(AllowableEncryptionValues...).Error(ErrInvalidEncryptionValueMsg),
	)
	if err != nil {
		return err
	}
	return nil
}

// https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Authentication
func (t *RealVNCServerTask) ValidateAuthenticationField() error {
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
func (t *RealVNCServerTask) ValidatePermissionsField() error {
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
func (t *RealVNCServerTask) ValidateLogField() error {
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
func (t *RealVNCServerTask) ValidateCaptureMethodField() error {
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
