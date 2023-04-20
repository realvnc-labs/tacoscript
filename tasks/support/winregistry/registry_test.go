//go:build windows
// +build windows

package winregistry_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"

	"github.com/realvnc-labs/tacoscript/tasks/support/winregistry"
)

const baseTestKey = `HKLM:\Software\TestTacoScript`
const testKey = `UnitTestRun`

func setup(t *testing.T) {
	t.Helper()

	err := winregistry.DeleteKeyRecursive(`HKLM:\Software\TestTacoScript`)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		require.NoError(t, err)
	}
}

func newTestKeyPath(key string) (testKeyPath string) {
	return baseTestKey + `\` + key
}

func TestShouldGetStringValue(t *testing.T) {
	// assumes golang is install in the default windows location
	keyPath := `HKCU:\Software\GoProgrammingLanguage`
	name := `installLocation`

	found, val, err := winregistry.GetValue(keyPath, name, winregistry.REG_SZ)

	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, `C:\Program Files\Go\`, val)
}

func TestShouldEnsureNewRegistryValueIsPresent(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	val := `123456789`
	valType := winregistry.REG_SZ

	setup(t)

	updated, desc, err := winregistry.SetValue(keyPath, name, val, valType)
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, "added new key", desc)

	found, currVal, err := winregistry.GetValue(keyPath, name, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, currVal)
}

func TestShouldEnsureExistingRegistryValueIsPresent(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	val := "1234567890"
	valType := winregistry.REG_SZ

	setup(t)

	// set initial value
	winregistry.SetValue(keyPath, name, val, valType)

	// now update again without no change
	updated, desc, err := winregistry.SetValue(keyPath, name, val, valType)
	require.NoError(t, err)

	assert.False(t, updated)
	assert.Equal(t, "matching existing value", desc)

	found, currVal, err := winregistry.GetValue(keyPath, name, valType)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, currVal)
}

func TestShouldEnsureExistingRegistryValueIsUpdated(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	val := `123456789`
	valType := winregistry.REG_SZ

	setup(t)

	// set an initial value
	winregistry.SetValue(keyPath, name, val+"abc", valType)

	// now update again
	updated, desc, err := winregistry.SetValue(keyPath, name, val, valType)
	assert.NoError(t, err)

	// new value will have updated as true as the value should have been updated
	assert.True(t, updated)
	assert.Equal(t, "existing value updated", desc)

	_, updatedVal, err := winregistry.GetValue(keyPath, name, winregistry.REG_SZ)
	require.NoError(t, err)

	assert.Equal(t, val, updatedVal)
}

func TestShouldEnsureExistingRegistryValueIsUpdatedWhenTypeChange(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	var val uint32 = 1
	valType := winregistry.REG_DWORD

	setup(t)

	winregistry.SetValue(keyPath, name, "existing value", winregistry.REG_SZ)

	updated, desc, err := winregistry.SetValue(keyPath, name, val, valType)

	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, "existing value updated", desc)

	found, currVal, err := winregistry.GetValue(keyPath, name, valType)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, currVal)
}

func TestShouldEnsureExistingRegistryValueIsAbsent(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`

	setup(t)

	winregistry.SetValue(keyPath, name, "value to be removed", winregistry.REG_SZ)

	updated, desc, err := winregistry.RemoveValue(keyPath, name)
	assert.NoError(t, err)

	// updated set to true indicates that a value was removed
	assert.True(t, updated)
	assert.Equal(t, "value removed", desc)

	found, _, err := winregistry.GetValue(keyPath, name, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestShouldWhenAbsentOnlyRemoveValue(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	altName := name + "_alt"

	setup(t)

	winregistry.SetValue(keyPath, altName, "value to remain", winregistry.REG_SZ)

	updated, desc, err := winregistry.RemoveValue(keyPath, name)
	require.NoError(t, err)

	assert.False(t, updated)
	assert.Equal(t, "no existing value", desc)

	found, _, err := winregistry.GetValue(keyPath, name, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)

	found, val, err := winregistry.GetValue(keyPath, altName, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value to remain", val)
}

func TestShouldWhenAbsentAndNoExistingValueHaveCorrectDescription(t *testing.T) {
	keyPath := newTestKeyPath(testKey)
	name := `testValue`
	altName := name + "_alt"

	setup(t)

	winregistry.SetValue(keyPath, altName, "value to remain", winregistry.REG_SZ)

	updated, desc, err := winregistry.RemoveValue(keyPath, name)
	require.NoError(t, err)

	assert.False(t, updated)
	assert.Equal(t, "no existing value", desc)
}

func TestShouldEnsureExistingRegistryKeyIsAbsent(t *testing.T) {
	keyPathToRemove := newTestKeyPath(testKey)
	name := `testValue`
	altName := name + "_alt"

	setup(t)

	winregistry.SetValue(keyPathToRemove, name, "1234", winregistry.REG_SZ)
	winregistry.SetValue(keyPathToRemove, altName, "value to remove also", winregistry.REG_SZ)

	updated, desc, err := winregistry.RemoveKey(keyPathToRemove)
	assert.NoError(t, err)

	// updated set to true indicates that a value was removed
	assert.True(t, updated)
	assert.Equal(t, "key removed", desc)

	found, _, err := winregistry.GetValue(keyPathToRemove, name, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)

	found, _, err = winregistry.GetValue(keyPathToRemove, altName, winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestShouldDeleteSubKeyRecursively(t *testing.T) {
	keyPath := createTestRegBranch(t)

	updated, desc, err := winregistry.RemoveKey(keyPath + `\2`)
	assert.NoError(t, err)

	// updated set to true indicates that a value was removed
	assert.True(t, updated)
	assert.Equal(t, "key removed", desc)

	found, _, err := winregistry.GetValue(keyPath+`\2`, "2", winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)

	found, _, err = winregistry.GetValue(keyPath+`\6`, "4", winregistry.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
}

func TestShouldDeleteSubKeyWithoutChildrenRecursively(t *testing.T) {
	keyPath := createTestRegBranch(t)

	updated, desc, err := winregistry.RemoveKey(keyPath + `\6`)
	assert.NoError(t, err)

	// updated set to true indicates that a value was removed
	assert.True(t, updated)
	assert.Equal(t, "key removed", desc)

	found, _, err := winregistry.GetValue(keyPath+`\6`, "4", winregistry.REG_SZ)
	require.NoError(t, err)
	assert.False(t, found)

	found, _, err = winregistry.GetValue(keyPath+`\2`, "4", winregistry.REG_SZ)
	require.NoError(t, err)
	assert.True(t, found)
}

func createTestRegBranch(t *testing.T) (keyPath string) {
	keyPath = newTestKeyPath(testKey)

	createBranchLeaves(t, keyPath, 0, 9)
	createBranchLeaves(t, keyPath+`\2`, 2, 5)
	createBranchLeaves(t, keyPath+`\2\2`, 5, 7)
	createBranchLeaves(t, keyPath+`\2\2\3`, 2, 3)
	createBranchLeaves(t, keyPath+`\2\2\3\4`, 2, 3)
	createBranchLeaves(t, keyPath+`\2\2\3\5`, 2, 3)
	createBranchLeaves(t, keyPath+`\2\2\4`, 1, 7)
	createBranchLeaves(t, keyPath+`\6`, 4, 8)

	return keyPath
}

func createBranchLeaves(t *testing.T, keyPath string, from int, to int) {
	for i := from; i <= to; i++ {
		_, _, err := winregistry.SetValue(keyPath, strconv.Itoa(i), strconv.Itoa(i), winregistry.REG_SZ)
		require.NoError(t, err)
	}
}
