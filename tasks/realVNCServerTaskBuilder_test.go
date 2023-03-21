package tasks

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealVNCServerTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *RealVNCServerTask
		expectedError string
	}{
		{
			typeName: "realVNCServerType",
			path:     "realVNCServerPath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: EncryptionField, Value: "AlwaysOn"}},
				yaml.MapSlice{yaml.MapItem{Key: AuthenticationField, Value: "SingleSignOn+Radius,SystemAuth+Radius"}},
				yaml.MapSlice{yaml.MapItem{Key: PermissionsField, Value: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryConnectField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryOnlyIfLoggedOnField, Value: "false"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryConnectTimeoutSecsField, Value: "60"}},
				yaml.MapSlice{yaml.MapItem{Key: BlankScreenField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: ConnNotifyTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: ConnNotifyAlwaysField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: IdleTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: LogField, Value: "*:file:10,Connections:file:100"}},
				yaml.MapSlice{yaml.MapItem{Key: CaptureMethodField, Value: "1"}},

				yaml.MapSlice{yaml.MapItem{Key: ConfigFileField, Value: "/tmp/config_file.conf"}},
				yaml.MapSlice{yaml.MapItem{Key: ServerModeField, Value: "service"}},

				yaml.MapSlice{yaml.MapItem{Key: ExecPathField, Value: "test/path"}},
				yaml.MapSlice{yaml.MapItem{Key: ExecCmdField, Value: "test_cmd"}},
				yaml.MapSlice{yaml.MapItem{Key: SkipReloadField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: BackupExtensionField, Value: "orig"}},
				yaml.MapSlice{yaml.MapItem{Key: SkipBackupField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: ShellField, Value: "someshell"}},
			},
			expectedTask: &RealVNCServerTask{
				TypeName: "realVNCServerType",
				Path:     "realVNCServerPath",

				Encryption:              "AlwaysOn",
				Authentication:          "SingleSignOn+Radius,SystemAuth+Radius",
				Permissions:             "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				QueryConnect:            true,
				QueryOnlyIfLoggedOn:     false,
				QueryConnectTimeoutSecs: 60,
				BlankScreen:             true,
				ConnNotifyTimeoutSecs:   30,
				ConnNotifyAlways:        true,
				IdleTimeoutSecs:         30,
				Log:                     "*:file:10,Connections:file:100",
				CaptureMethod:           1,
				ConfigFile:              "/tmp/config_file.conf",
				ServerMode:              "Service",
				Creates:                 []string{"/tmp/creates-file.txt"},
				OnlyIf:                  []string{"/tmp/onlyif-file.txt"},
				Unless: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
				Require: []string{"/tmp/required-file.txt"},

				Backup:     "orig",
				ExecPath:   "test/path",
				ExecCmd:    "test_cmd",
				SkipReload: true,
				SkipBackup: true,

				Shell: "someshell",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			taskBuilder := RealVNCServerTaskBuilder{}
			task, err := taskBuilder.Build(
				tc.typeName,
				tc.path,
				tc.values,
			)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, tc.expectedError)
				return
			}
			require.NoError(t, err)

			actualTask, ok := task.(*RealVNCServerTask)
			require.True(t, ok)

			assertRealVNCServerTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func TestRealVNCServerTaskBuilderWithUnsets(t *testing.T) {
	testCases := []struct {
		name          string
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *RealVNCServerTask
		expectedError string
	}{
		{
			name:     "with unsets",
			typeName: "realVNCServerType",
			path:     "realVNCServerPath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: EncryptionField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: AuthenticationField, Value: "SingleSignOn+Radius,SystemAuth+Radius"}},
				yaml.MapSlice{yaml.MapItem{Key: PermissionsField, Value: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryConnectField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryOnlyIfLoggedOnField, Value: "false"}},
				yaml.MapSlice{yaml.MapItem{Key: QueryConnectTimeoutSecsField, Value: "60"}},
				yaml.MapSlice{yaml.MapItem{Key: BlankScreenField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: ConnNotifyTimeoutSecsField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: ConnNotifyAlwaysField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: IdleTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: LogField, Value: "*:file:10,Connections:file:100"}},
				yaml.MapSlice{yaml.MapItem{Key: CaptureMethodField, Value: "1"}},
				yaml.MapSlice{yaml.MapItem{Key: ConfigFileField, Value: "/tmp/config_file.conf"}},
				yaml.MapSlice{yaml.MapItem{Key: ServerModeField, Value: "service"}},

				yaml.MapSlice{yaml.MapItem{Key: CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: UnlessField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: ShellField, Value: "someshell"}},
			},
			expectedTask: &RealVNCServerTask{
				TypeName: "realVNCServerType",
				Path:     "realVNCServerPath",

				Encryption:              "",
				Authentication:          "SingleSignOn+Radius,SystemAuth+Radius",
				Permissions:             "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				QueryConnect:            true,
				QueryOnlyIfLoggedOn:     false,
				QueryConnectTimeoutSecs: 60,
				BlankScreen:             false,
				ConnNotifyTimeoutSecs:   0,
				ConnNotifyAlways:        true,
				IdleTimeoutSecs:         30,
				Log:                     "*:file:10,Connections:file:100",
				CaptureMethod:           1,
				ConfigFile:              "/tmp/config_file.conf",
				ServerMode:              "Service",
				Creates:                 []string{"/tmp/creates-file.txt"},
				OnlyIf:                  []string{"/tmp/onlyif-file.txt"},
				Unless: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
				Require: []string{"/tmp/required-file.txt"},

				Shell: "someshell",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			taskBuilder := RealVNCServerTaskBuilder{}
			task, err := taskBuilder.Build(
				tc.typeName,
				tc.path,
				tc.values,
			)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, tc.expectedError)
				return
			}
			require.NoError(t, err)

			actualTask, ok := task.(*RealVNCServerTask)
			require.True(t, ok)

			assertRealVNCServerTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertRealVNCServerTaskEquals(t *testing.T, expectedTask, actualTask *RealVNCServerTask) {
	assert.Equal(t, expectedTask.Path, actualTask.GetPath())

	assert.Equal(t, expectedTask.Encryption, actualTask.Encryption)
	assert.Equal(t, expectedTask.Authentication, actualTask.Authentication)
	assert.Equal(t, expectedTask.Permissions, actualTask.Permissions)
	assert.Equal(t, expectedTask.QueryConnect, actualTask.QueryConnect)
	assert.Equal(t, expectedTask.QueryOnlyIfLoggedOn, actualTask.QueryOnlyIfLoggedOn)
	assert.Equal(t, expectedTask.QueryConnectTimeoutSecs, actualTask.QueryConnectTimeoutSecs)
	assert.Equal(t, expectedTask.BlankScreen, actualTask.BlankScreen)
	assert.Equal(t, expectedTask.ConnNotifyTimeoutSecs, actualTask.ConnNotifyTimeoutSecs)
	assert.Equal(t, expectedTask.ConnNotifyAlways, actualTask.ConnNotifyAlways)
	assert.Equal(t, expectedTask.IdleTimeoutSecs, actualTask.IdleTimeoutSecs)
	assert.Equal(t, expectedTask.Log, actualTask.Log)
	assert.Equal(t, expectedTask.CaptureMethod, actualTask.CaptureMethod)
	assert.Equal(t, expectedTask.ConfigFile, actualTask.ConfigFile)

	assert.Equal(t, expectedTask.Require, actualTask.Require)

	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
}
