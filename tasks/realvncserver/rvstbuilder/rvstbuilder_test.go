package rvstbuilder

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/realvnc-labs/tacoscript/tasks"
	"github.com/realvnc-labs/tacoscript/tasks/realvncserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *realvncserver.RvsTask
		expectedError string
	}{
		{
			typeName: "realVNCServerType",
			path:     "realVNCServerPath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.EncryptionField, Value: "AlwaysOn"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.AuthenticationField, Value: "SingleSignOn+Radius,SystemAuth+Radius"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.PermissionsField, Value: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryConnectField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryOnlyIfLoggedOnField, Value: "false"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryConnectTimeoutSecsField, Value: "60"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BlankScreenField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConnNotifyTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConnNotifyAlwaysField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.IdleTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.LogField, Value: "*:file:10,Connections:file:100"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CaptureMethodField, Value: "1"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: "/tmp/config_file.conf"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.ReloadExecPathField, Value: "test/path/exe_cmd"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipReloadField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.BackupExtensionField, Value: "orig"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.SkipBackupField, Value: true}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
			},
			expectedTask: &realvncserver.RvsTask{
				TypeName: "realVNCServerType",
				Path:     "realVNCServerPath",

				Encryption:          "AlwaysOn",
				Authentication:      "SingleSignOn+Radius,SystemAuth+Radius",
				Permissions:         "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				QueryConnect:        true,
				QueryOnlyIfLoggedOn: false,
				QueryConnectTimeout: 60,
				BlankScreen:         true,
				ConnNotifyTimeout:   30,
				ConnNotifyAlways:    true,
				IdleTimeout:         30,
				Log:                 "*:file:10,Connections:file:100",
				CaptureMethod:       1,
				ConfigFile:          "/tmp/config_file.conf",
				ServerMode:          "Service",
				Creates:             []string{"/tmp/creates-file.txt"},
				OnlyIf:              []string{"/tmp/onlyif-file.txt"},
				Unless: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
				Require: []string{"/tmp/required-file.txt"},

				Backup:         "orig",
				ReloadExecPath: "test/path/exe_cmd",
				SkipReload:     true,
				SkipBackup:     true,

				Shell: "someshell",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			taskBuilder := TaskBuilder{}
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

			actualTask, ok := task.(*realvncserver.RvsTask)
			require.True(t, ok)

			assertRealVNCServerTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func TestTaskBuilderWithUnsets(t *testing.T) {
	testCases := []struct {
		name          string
		typeName      string
		path          string
		values        []interface{}
		expectedTask  *realvncserver.RvsTask
		expectedError string
	}{
		{
			name:     "with unsets",
			typeName: "realVNCServerType",
			path:     "realVNCServerPath",
			values: []interface{}{
				yaml.MapSlice{yaml.MapItem{Key: tasks.EncryptionField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.AuthenticationField, Value: "SingleSignOn+Radius,SystemAuth+Radius"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.PermissionsField, Value: "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryConnectField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryOnlyIfLoggedOnField, Value: "false"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.QueryConnectTimeoutSecsField, Value: "60"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.BlankScreenField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConnNotifyTimeoutSecsField, Value: "!UNSET!"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConnNotifyAlwaysField, Value: "true"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.IdleTimeoutSecsField, Value: "30"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.LogField, Value: "*:file:10,Connections:file:100"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.CaptureMethodField, Value: "1"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ConfigFileField, Value: "/tmp/config_file.conf"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.ServerModeField, Value: "Service"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.CreatesField, Value: "/tmp/creates-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.OnlyIfField, Value: "/tmp/onlyif-file.txt"}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.UnlessField, Value: []interface{}{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				}}},
				yaml.MapSlice{yaml.MapItem{Key: tasks.RequireField, Value: "/tmp/required-file.txt"}},

				yaml.MapSlice{yaml.MapItem{Key: tasks.ShellField, Value: "someshell"}},
			},
			expectedTask: &realvncserver.RvsTask{
				TypeName: "realVNCServerType",
				Path:     "realVNCServerPath",

				Encryption:          "",
				Authentication:      "SingleSignOn+Radius,SystemAuth+Radius",
				Permissions:         "superuser:f,%vncusers:d,johndoe:v,janedoe:skp-t!r",
				QueryConnect:        true,
				QueryOnlyIfLoggedOn: false,
				QueryConnectTimeout: 60,
				BlankScreen:         false,
				ConnNotifyTimeout:   0,
				ConnNotifyAlways:    true,
				IdleTimeout:         30,
				Log:                 "*:file:10,Connections:file:100",
				CaptureMethod:       1,
				ConfigFile:          "/tmp/config_file.conf",
				ServerMode:          "Service",
				Creates:             []string{"/tmp/creates-file.txt"},
				OnlyIf:              []string{"/tmp/onlyif-file.txt"},
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
			taskBuilder := TaskBuilder{}
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

			actualTask, ok := task.(*realvncserver.RvsTask)
			require.True(t, ok)

			assertRealVNCServerTaskEquals(t, tc.expectedTask, actualTask)
		})
	}
}

func assertRealVNCServerTaskEquals(t *testing.T, expectedTask, actualTask *realvncserver.RvsTask) {
	assert.Equal(t, expectedTask.Path, actualTask.GetPath())

	assert.Equal(t, expectedTask.Encryption, actualTask.Encryption)
	assert.Equal(t, expectedTask.Authentication, actualTask.Authentication)
	assert.Equal(t, expectedTask.Permissions, actualTask.Permissions)
	assert.Equal(t, expectedTask.QueryConnect, actualTask.QueryConnect)
	assert.Equal(t, expectedTask.QueryOnlyIfLoggedOn, actualTask.QueryOnlyIfLoggedOn)
	assert.Equal(t, expectedTask.QueryConnectTimeout, actualTask.QueryConnectTimeout)
	assert.Equal(t, expectedTask.BlankScreen, actualTask.BlankScreen)
	assert.Equal(t, expectedTask.ConnNotifyTimeout, actualTask.ConnNotifyTimeout)
	assert.Equal(t, expectedTask.ConnNotifyAlways, actualTask.ConnNotifyAlways)
	assert.Equal(t, expectedTask.IdleTimeout, actualTask.IdleTimeout)
	assert.Equal(t, expectedTask.Log, actualTask.Log)
	assert.Equal(t, expectedTask.CaptureMethod, actualTask.CaptureMethod)

	assert.Equal(t, expectedTask.ConfigFile, actualTask.ConfigFile)
	assert.Equal(t, expectedTask.ServerMode, actualTask.ServerMode)
	assert.Equal(t, expectedTask.ReloadExecPath, actualTask.ReloadExecPath)
	assert.Equal(t, expectedTask.SkipReload, actualTask.SkipReload)
	assert.Equal(t, expectedTask.UseVNCLicenseReload, actualTask.UseVNCLicenseReload)
	assert.Equal(t, expectedTask.Backup, actualTask.Backup)
	assert.Equal(t, expectedTask.SkipBackup, actualTask.SkipBackup)
	assert.Equal(t, expectedTask.Require, actualTask.Require)

	assert.Equal(t, expectedTask.Creates, actualTask.Creates)
	assert.Equal(t, expectedTask.OnlyIf, actualTask.OnlyIf)
	assert.Equal(t, expectedTask.Unless, actualTask.Unless)
}
