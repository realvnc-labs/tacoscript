package realvncserver_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/realvnc-labs/tacoscript/tasks/shared/fieldstatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func WithNameMap(fieldKey string, fieldName string) (change fieldstatus.NameMap) {
	return fieldstatus.NameMap{
		fieldKey: fieldName,
	}
}

func newTrackerWithSingleFieldStatus(fieldKey string, fieldName string) (tracker *fieldstatus.FieldNameStatusTracker) {
	tracker = fieldstatus.NewFieldNameStatusTrackerWithMapAndStatus(
		WithNameMap(fieldKey, fieldName),
		fieldstatus.StatusMap{
			fieldName: fieldstatus.FieldStatus{HasNewValue: true},
		})

	return tracker
}

func initMapperTracker(task *realvncserver.RvsTask) {
	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")
	task.SetMapper(tracker)
	task.SetTracker(tracker)
}

func TestRealVNCNameFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "valid name value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/tmp/config.conf",
			},
		},
		{
			name: "invalid path value",
			task: realvncserver.RvsTask{
				// Path: "MyTask",
				ConfigFile: "/tmp/config.conf",
			},
			expectedError: fmt.Errorf("empty required value at path '.%s'", tasks.NameField).Error(),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCConfigFileBaseFieldValidations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	testCases := []struct {
		name               string
		task               realvncserver.RvsTask
		goos               string
		expectedErrorMsg   string
		expectedConfigFile string
	}{
		{
			name: "valid config_file value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
			},
			goos:               "any",
			expectedConfigFile: "/config/file/name/here",
		},
		{
			name: "when no config file, use service server mode config file",
			task: realvncserver.RvsTask{
				Path: "MyTask",
			},
			goos:               "any",
			expectedConfigFile: realvncserver.DefaultServiceServerModeConfigFile,
		},
		{
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.ServiceServerMode,
			},
			goos:               "any",
			expectedConfigFile: realvncserver.DefaultServiceServerModeConfigFile,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			goos := runtime.GOOS
			if tc.goos != "any" {
				goos = tc.goos
			}

			err := task.Validate(goos)

			if tc.expectedErrorMsg != "" {
				assert.EqualError(t, err, tc.expectedErrorMsg)
				return
			}

			assert.Contains(t, task.ConfigFile, tc.expectedConfigFile)
			require.NoError(t, err)
		})
	}
}

func TestRealVNCConfigFileExtendedFieldValidations(t *testing.T) {
	// TODO: (rs): remove this Skip when support for user and virtual server modes is reintroduced.
	t.Skip()

	if runtime.GOOS == "windows" {
		t.Skip()
	}
	testCases := []struct {
		name               string
		task               realvncserver.RvsTask
		goos               string
		expectedErrorMsg   string
		expectedConfigFile string
	}{
		{
			name: "default path when user server mode",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.UserServerMode,
			},
			goos:               "any",
			expectedConfigFile: realvncserver.DefaultUserServerModeConfigFile,
		},
		{
			name: "default path when virtual server mode",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.VirtualServerMode,
			},
			goos:               "linux",
			expectedConfigFile: realvncserver.DefaultVirtualServerModeConfigFile,
		},
		{
			name: "error when virtual server mode and darwin",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.VirtualServerMode,
			},
			goos:               "darwin",
			expectedConfigFile: realvncserver.DefaultVirtualServerModeConfigFile,
			expectedErrorMsg:   realvncserver.ErrServerModeCannotBeVirtualWhenNotLinuxMsg,
		},
		{
			name: "error when virtual server mode and windows",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.VirtualServerMode,
			},
			goos:             "windows",
			expectedErrorMsg: realvncserver.ErrServerModeCannotBeVirtualWhenNotLinuxMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			goos := runtime.GOOS
			if tc.goos != "any" {
				goos = tc.goos
			}

			err := task.Validate(goos)

			if tc.expectedErrorMsg != "" {
				assert.EqualError(t, err, tc.expectedErrorMsg)
				return
			}

			assert.Contains(t, task.ConfigFile, tc.expectedConfigFile)
			require.NoError(t, err)
		})
	}
}

func TestShouldSetUseVNCLicenseReloadWhenVirtualServiceMode(t *testing.T) {
	// TODO: (rs): remove this Skip when support for user and virtual server modes is reintroduced.
	t.Skip()

	cases := []struct {
		name                  string
		task                  realvncserver.RvsTask
		goos                  string
		expectedLicenseReload bool
	}{
		{
			name: "virtual mode license reload linux",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ServerMode: realvncserver.VirtualServerMode,
			},
			goos:                  "linux",
			expectedLicenseReload: true,
		},
		{
			name: "no license reload darwin",
			task: realvncserver.RvsTask{
				Path: "MyTask",
			},
			goos:                  "darwin",
			expectedLicenseReload: false,
		},
		{
			name: "no license reload windows",
			task: realvncserver.RvsTask{
				Path: "MyTask",
			},
			goos:                  "windows",
			expectedLicenseReload: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			err := task.Validate("linux")
			require.NoError(t, err)
			assert.Equal(t, tc.expectedLicenseReload, task.UseVNCLicenseReload)
		})
	}
}

func TestShouldNotSetUseVNCLicenseReloadWhenNotVirtualServiceMode(t *testing.T) {
	// TODO: (rs): remove this Skip when support for user and virtual server modes is reintroduced.
	t.Skip()

	task := realvncserver.RvsTask{
		Path: "MyTask",
	}

	initMapperTracker(&task)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.False(t, task.UseVNCLicenseReload)
}

func TestShouldErrorWhenVirtualServerModeAndNotLinux(t *testing.T) {
	// TODO: (rs): remove this Skip when support for user and virtual server modes is reintroduced.
	t.Skip()

	task := &realvncserver.RvsTask{
		Path:       "MyTask",
		ServerMode: realvncserver.VirtualServerMode,
	}

	err := task.ValidateServerModeField("windows")
	require.ErrorContains(t, err, realvncserver.ErrServerModeCannotBeVirtualWhenNotLinuxMsg)
	err = task.ValidateServerModeField("darwin")
	require.ErrorContains(t, err, realvncserver.ErrServerModeCannotBeVirtualWhenNotLinuxMsg)
}

func TestShouldNotErrorWhenVirtualServerModeAndLinux(t *testing.T) {
	// TODO: (rs): remove this Skip when support for user and virtual server modes is reintroduced.
	t.Skip()

	task := &realvncserver.RvsTask{
		Path:       "MyTask",
		ServerMode: realvncserver.VirtualServerMode,
	}

	err := task.ValidateServerModeField("linux")
	require.NoError(t, err)
}

func TestRealVNCServerEncryptionFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "valid encryption value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "AlwaysOn",
			},
		},
		{
			name: "invalid encryption value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "randomvalue",
			},
			expectedError: realvncserver.ErrInvalidEncryptionValueMsg,
		},
		{
			name: "invalid encryption value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "AlwaysOn   # this comment is an error",
			},
			expectedError: realvncserver.ErrInvalidEncryptionValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			initMapperTracker(&task)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCServerAuthenticationFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "invalid authentication value",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "invalidValue",
			},
			expectedError: realvncserver.ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "valid authentication value",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,SystemAuth+Radius",
			},
		},
		{
			name: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
			},
		},
		{
			name: "missing additional authentication",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,",
			},
			expectedError: realvncserver.ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "missing additional authentication scheme",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+",
			},
			expectedError: realvncserver.ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "contains illegal comment",
			task: realvncserver.RvsTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn    # not allowed comment",
			},
			expectedError: realvncserver.ErrInvalidAuthenticationValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			tracker := newTrackerWithSingleFieldStatus("authentication", "Authentication")
			task.SetMapper(tracker)
			task.SetTracker(tracker)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCServerPermissionsFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "invalid permissions value",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "invalidValue",
			},
			expectedError: realvncserver.ErrInvalidPermisssionsMsg,
		},
		{
			name: "valid permissions value",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
			},
		},
		{
			name: "valid permissions value - user with no permissions",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:,%vncusers:d,johndoe:v,janedoe:skp-t!r",
			},
		},
		{
			name: "permissions with whitespace",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, %vncusers :d , johndoe:v, janedoe:skp-t!r",
			},
		},
		{
			name: "missing additional permissions",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, ",
			},
			expectedError: realvncserver.ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :fx, ",
			},
			expectedError: realvncserver.ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character - has space",
			task: realvncserver.RvsTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f x",
			},
			expectedError: realvncserver.ErrInvalidPermisssionsMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task
			tracker := newTrackerWithSingleFieldStatus("permissions", "Permissions")
			task.SetMapper(tracker)
			task.SetTracker(tracker)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCServerLogsFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "invalid value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "invalidValue",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "valid value",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:file:10,Connections:file:100",
			},
		},
		{
			name: "missing log area",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ":file:10,Connections:file:100",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log target",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*::10,Connections:file:100",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log level",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:,Connections:file:100",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 1",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:10,",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 2",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ",*:stderr:10",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - value not permitted",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:11",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too high",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:1000",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too low",
			task: realvncserver.RvsTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:-100",
			},
			expectedError: realvncserver.ErrInvalidLogsValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			tracker := newTrackerWithSingleFieldStatus("log", "Log")
			task.SetMapper(tracker)
			task.SetTracker(tracker)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCServerCaptureMethodFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          realvncserver.RvsTask
		expectedError string
	}{
		{
			name: "valid value",
			task: realvncserver.RvsTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 1,
			},
		},
		{
			name: "negative",
			task: realvncserver.RvsTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: -1,
			},
			expectedError: realvncserver.ErrInvalidCaptureMethodValueMsg,
		},
		{
			name: "too high",
			task: realvncserver.RvsTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 100,
			},
			expectedError: realvncserver.ErrInvalidCaptureMethodValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

			tracker := newTrackerWithSingleFieldStatus("capture_method", "CaptureMethod")
			task.SetMapper(tracker)
			task.SetTracker(tracker)

			err := task.Validate(runtime.GOOS)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestShouldSetDefaultBackupExtension(t *testing.T) {
	task := &realvncserver.RvsTask{
		Path:       "MyTask",
		ConfigFile: "/config/file/name/here",
	}

	tracker := fieldstatus.NewFieldNameStatusTracker()
	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.Equal(t, "bak", task.Backup)
}

func TestShouldSetBackupExtension(t *testing.T) {
	task := &realvncserver.RvsTask{
		Path:       "MyTask",
		ConfigFile: "/config/file/name/here",
		Backup:     "orig",
	}

	tracker := fieldstatus.NewFieldNameStatusTracker()
	task.SetMapper(tracker)
	task.SetTracker(tracker)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.Equal(t, "orig", task.Backup)
}
