package realvncserver

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func WithNameMap(fieldKey string, fieldName string) (change tasks.FieldNameMap) {
	return tasks.FieldNameMap{
		fieldKey: fieldName,
	}
}

func newTrackerWithSingleFieldStatus(fieldKey string, fieldName string) (tracker *tasks.FieldNameStatusTracker) {
	return &tasks.FieldNameStatusTracker{
		NameMap: WithNameMap(fieldKey, fieldName),
		StatusMap: tasks.FieldStatusMap{
			fieldName: tasks.FieldStatus{HasNewValue: true},
		},
	}
}

func initMapperTracker(task *RealVNCServerTask) {
	tracker := newTrackerWithSingleFieldStatus("encryption", "Encryption")
	task.Mapper = tracker
	task.Tracker = tracker
}

func TestRealVNCNameFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid name value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/tmp/config.conf",
			},
		},
		{
			name: "invalid path value",
			task: RealVNCServerTask{
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

func TestRealVNCConfigFileFieldValidations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	testCases := []struct {
		name               string
		task               RealVNCServerTask
		goos               string
		expectedErrorMsg   string
		expectedConfigFile string
	}{
		{
			name: "valid config_file value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
			},
			goos:               "any",
			expectedConfigFile: "/config/file/name/here",
		},
		{
			name: "when no config file, use service server mode config file",
			task: RealVNCServerTask{
				Path: "MyTask",
			},
			goos:               "any",
			expectedConfigFile: DefaultServiceServerModeConfigFile,
		},
		{
			name: "default path when user server mode",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: UserServerMode,
			},
			goos:               "any",
			expectedConfigFile: DefaultUserServerModeConfigFile,
		},
		{
			name: "default path when service server mode",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: ServiceServerMode,
			},
			goos:               "any",
			expectedConfigFile: DefaultServiceServerModeConfigFile,
		},
		{
			name: "default path when virtual server mode",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: VirtualServerMode,
			},
			goos:               "linux",
			expectedConfigFile: DefaultVirtualServerModeConfigFile,
		},
		{
			name: "error when virtual server mode and darwin",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: VirtualServerMode,
			},
			goos:               "darwin",
			expectedConfigFile: DefaultVirtualServerModeConfigFile,
			expectedErrorMsg:   ErrServerModeCannotBeVirtualWhenNotLinuxMsg,
		},
		{
			name: "error when virtual server mode and windows",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: VirtualServerMode,
			},
			goos:             "windows",
			expectedErrorMsg: ErrServerModeCannotBeVirtualWhenNotLinuxMsg,
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
	cases := []struct {
		name                  string
		task                  RealVNCServerTask
		goos                  string
		expectedLicenseReload bool
	}{
		{
			name: "virtual mode license reload linux",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ServerMode: VirtualServerMode,
			},
			goos:                  "linux",
			expectedLicenseReload: true,
		},
		{
			name: "no license reload darwin",
			task: RealVNCServerTask{
				Path: "MyTask",
			},
			goos:                  "darwin",
			expectedLicenseReload: false,
		},
		{
			name: "no license reload windows",
			task: RealVNCServerTask{
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
	task := RealVNCServerTask{
		Path: "MyTask",
	}

	initMapperTracker(&task)

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.False(t, task.UseVNCLicenseReload)
}

func TestShouldErrorWhenVirtualServerModeAndNotLinux(t *testing.T) {
	task := &RealVNCServerTask{
		Path:       "MyTask",
		ServerMode: VirtualServerMode,
	}

	err := task.ValidateServerModeField("windows")
	require.ErrorContains(t, err, ErrServerModeCannotBeVirtualWhenNotLinuxMsg)
	err = task.ValidateServerModeField("darwin")
	require.ErrorContains(t, err, ErrServerModeCannotBeVirtualWhenNotLinuxMsg)
}

func TestShouldNotErrorWhenVirtualServerModeAndLinux(t *testing.T) {
	task := &RealVNCServerTask{
		Path:       "MyTask",
		ServerMode: VirtualServerMode,
	}

	err := task.ValidateServerModeField("linux")
	require.NoError(t, err)
}

func TestRealVNCServerEncryptionFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid encryption value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "AlwaysOn",
			},
		},
		{
			name: "invalid encryption value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "randomvalue",
			},
			expectedError: ErrInvalidEncryptionValueMsg,
		},
		{
			name: "invalid encryption value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "AlwaysOn   # this comment is an error",
			},
			expectedError: ErrInvalidEncryptionValueMsg,
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
	tracker := newTrackerWithSingleFieldStatus("authentication", "Authentication")

	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid authentication value",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "invalidValue",
				Mapper:         tracker,
				Tracker:        tracker,
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "valid authentication value",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,SystemAuth+Radius",
				Mapper:         tracker,
				Tracker:        tracker,
			},
		},
		{
			name: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
				Mapper:         tracker,
				Tracker:        tracker,
			},
		},
		{
			name: "missing additional authentication",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,",
				Mapper:         tracker,
				Tracker:        tracker,
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "missing additional authentication scheme",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+",
				Mapper:         tracker,
				Tracker:        tracker,
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "contains illegal comment",
			task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn    # not allowed comment",
				Mapper:         tracker,
				Tracker:        tracker,
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

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
	tracker := newTrackerWithSingleFieldStatus("permissions", "Permissions")

	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid permissions value",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "invalidValue",
				Mapper:      tracker,
				Tracker:     tracker,
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "valid permissions value",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				Mapper:      tracker,
				Tracker:     tracker,
			},
		},
		{
			name: "valid permissions value - user with no permissions",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				Mapper:      tracker,
				Tracker:     tracker,
			},
		},
		{
			name: "permissions with whitespace",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, %vncusers :d , johndoe:v, janedoe:skp-t!r",
				Mapper:      tracker,
				Tracker:     tracker,
			},
		},
		{
			name: "missing additional permissions",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, ",
				Mapper:      tracker,
				Tracker:     tracker,
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :fx, ",
				Mapper:      tracker,
				Tracker:     tracker,
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character - has space",
			task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f x",
				Mapper:      tracker,
				Tracker:     tracker,
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

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
	tracker := newTrackerWithSingleFieldStatus("log", "Log")

	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "invalidValue",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "valid value",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:file:10,Connections:file:100",
				Mapper:     tracker,
				Tracker:    tracker,
			},
		},
		{
			name: "missing log area",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ":file:10,Connections:file:100",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log target",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*::10,Connections:file:100",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log level",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:,Connections:file:100",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 1",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:10,",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 2",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ",*:stderr:10",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - value not permitted",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:11",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too high",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:1000",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too low",
			task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:-100",
				Mapper:     tracker,
				Tracker:    tracker,
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

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
	tracker := newTrackerWithSingleFieldStatus("capture_method", "CaptureMethod")

	testCases := []struct {
		name          string
		task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid value",
			task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 1,
				Mapper:        tracker,
				Tracker:       tracker,
			},
		},
		{
			name: "negative",
			task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: -1,
				Mapper:        tracker,
				Tracker:       tracker,
			},
			expectedError: ErrInvalidCaptureMethodValueMsg,
		},
		{
			name: "too high",
			task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 100,
				Mapper:        tracker,
				Tracker:       tracker,
			},
			expectedError: ErrInvalidCaptureMethodValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.task

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
	tracker := tasks.NewFieldCombinedTracker()
	task := &RealVNCServerTask{
		Path:       "MyTask",
		ConfigFile: "/config/file/name/here",
		Mapper:     tracker,
		Tracker:    tracker,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.Equal(t, "bak", task.Backup)
}

func TestShouldSetBackupExtension(t *testing.T) {
	tracker := tasks.NewFieldCombinedTracker()
	task := &RealVNCServerTask{
		Path:       "MyTask",
		ConfigFile: "/config/file/name/here",
		Backup:     "orig",
		Mapper:     tracker,
		Tracker:    tracker,
	}

	err := task.Validate(runtime.GOOS)
	require.NoError(t, err)

	assert.Equal(t, "orig", task.Backup)
}
