package tasks

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withNewValue(fieldKey string, fieldName string) (change map[string]FieldStatus) {
	return map[string]FieldStatus{
		fieldKey: {
			Name:        fieldName,
			HasNewValue: true,
		},
	}
}

func newTrackerWithSingleFieldStatus(fieldKey string, fieldName string) (tracker *FieldStatusTracker) {
	return &FieldStatusTracker{
		withNewValue(fieldKey, fieldName),
	}
}

func TestRealVNCNameFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid name value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/tmp/config.conf",
			},
		},
		{
			name: "invalid path value",
			Task: RealVNCServerTask{
				// Path: "MyTask",
				ConfigFile: "/tmp/config.conf",
			},
			expectedError: fmt.Errorf("empty required value at path '.%s'", NameField).Error(),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

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
		name          string
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid config_file value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
			},
		},
		{
			name: "invalid config_file field value",
			Task: RealVNCServerTask{
				Path: "MyTask",
			},
			expectedError: ErrConfigFileMustBeSpecifiedMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRealVNCServerEncryptionFieldValidations(t *testing.T) {
	testCases := []struct {
		name          string
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid encryption value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "AlwaysOn",
				tracker:    newTrackerWithSingleFieldStatus("encryption", "Encryption"),
			},
		},
		{
			name: "invalid encryption value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Encryption: "randomvalue",
				tracker:    newTrackerWithSingleFieldStatus("encryption", "Encryption"),
			},
			expectedError: ErrInvalidEncryptionValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

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
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid authentication value",
			Task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "invalidValue",
				tracker:        newTrackerWithSingleFieldStatus("authentication", "Authentication"),
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "valid authentication value",
			Task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,SystemAuth+Radius",
				tracker:        newTrackerWithSingleFieldStatus("authentication", "Authentication"),
			},
		},
		{
			name: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
			Task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn   +Radius   ,  SystemAuth+  Radius",
				tracker:        newTrackerWithSingleFieldStatus("authentication", "Authentication"),
			},
		},
		{
			name: "missing additional authentication",
			Task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+Radius,",
				tracker:        newTrackerWithSingleFieldStatus("authentication", "Authentication"),
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
		{
			name: "missing additional authentication scheme",
			Task: RealVNCServerTask{
				Path:           "MyTask",
				ConfigFile:     "/config/file/name/here",
				Authentication: "SingleSignOn+",
				tracker:        newTrackerWithSingleFieldStatus("authentication", "Authentication"),
			},
			expectedError: ErrInvalidAuthenticationValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

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
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid permissions value",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "invalidValue",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "valid permissions value",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
		},
		{
			name: "valid permissions value - user with no permissions",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser:,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
		},
		{
			name: "permissions with whitespace",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, %vncusers :d , johndoe:v, janedoe:skp-t!r",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
		},
		{
			name: "missing additional permissions",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f, ",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :fx, ",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
		{
			name: "invalid permissions character - has space",
			Task: RealVNCServerTask{
				Path:        "MyTask",
				ConfigFile:  "/config/file/name/here",
				Permissions: "superuser :f x",
				tracker:     newTrackerWithSingleFieldStatus("permissions", "Permissions"),
			},
			expectedError: ErrInvalidPermisssionsMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

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
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "invalid value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "invalidValue",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "valid value",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:file:10,Connections:file:100",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
		},
		{
			name: "missing log area",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ":file:10,Connections:file:100",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log target",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*::10,Connections:file:100",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "missing log level",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:,Connections:file:100",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 1",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:10,",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "incomplete value 2",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        ",*:stderr:10",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - value not permitted",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:11",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too high",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:1000",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
		{
			name: "invalid log level - too low",
			Task: RealVNCServerTask{
				Path:       "MyTask",
				ConfigFile: "/config/file/name/here",
				Log:        "*:stderr:-100",
				tracker:    newTrackerWithSingleFieldStatus("log", "Log"),
			},
			expectedError: ErrInvalidLogsValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

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
		Task          RealVNCServerTask
		expectedError string
	}{
		{
			name: "valid value",
			Task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 1,
				tracker:       newTrackerWithSingleFieldStatus("capture_method", "CaptureMethod"),
			},
		},
		{
			name: "negative",
			Task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: -1,
				tracker:       newTrackerWithSingleFieldStatus("capture_method", "CaptureMethod"),
			},
			expectedError: ErrInvalidCaptureMethodValueMsg,
		},
		{
			name: "too high",
			Task: RealVNCServerTask{
				Path:          "MyTask",
				ConfigFile:    "/config/file/name/here",
				CaptureMethod: 100,
				tracker:       newTrackerWithSingleFieldStatus("capture_method", "CaptureMethod"),
			},
			expectedError: ErrInvalidCaptureMethodValueMsg,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			task := tc.Task

			err := task.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestShouldSetDefaultBackupExtension(t *testing.T) {
	task := &RealVNCServerTask{
		Path:        "MyTask",
		ConfigFile:  "/config/file/name/here",
		Permissions: "superuser :fx, ",
		tracker:     newFieldStatusTracker(),
	}

	err := task.Validate()
	require.NoError(t, err)

	assert.Equal(t, "bak", task.Backup)
}

func TestShouldSetBackupExtension(t *testing.T) {
	task := &RealVNCServerTask{
		Path:       "MyTask",
		ConfigFile: "/config/file/name/here",
		tracker:    newFieldStatusTracker(),
		Backup:     "orig",
	}

	err := task.Validate()
	require.NoError(t, err)

	assert.Equal(t, "orig", task.Backup)
}
