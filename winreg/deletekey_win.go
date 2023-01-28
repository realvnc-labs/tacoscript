//go:build windows
// +build windows

package winreg

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// based on this implementation:
// https://blog.hackajob.co/how-to-do-windows-registry-fuctions-in-go/

func DeleteKeyRecursive(regPath string) error {
	rootKey, keyPath, err := getRootKey(regPath)
	if err != nil {
		return err
	}

	var access uint32 = registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS | registry.SET_VALUE
	k, err := registry.OpenKey(rootKey, keyPath, access)
	if err != nil {
		return err
	}

	defer k.Close()

	keyNames, err := k.ReadSubKeyNames(0)
	if err != nil {
		return fmt.Errorf("failed to get %q key from registry error: %v", regPath, err)
	}

	if err := deleteSubKeys(rootKey, keyPath, keyNames, access); err != nil {
		return err
	}

	// delete itself
	if err := registry.DeleteKey(k, ""); err != nil {
		return fmt.Errorf("cannot delete key path : %q error: %v", regPath, err)
	}
	return nil
}

func deleteSubKeys(rootKey registry.Key, rpath string, keyNames []string, access uint32) error {
	for _, keyName := range keyNames {
		keyPath := fmt.Sprintf("%s\\%s", rpath, keyName)

		k, err := registry.OpenKey(rootKey, keyPath, access)
		if err != nil {
			return fmt.Errorf("path %q not found on registry: %v", keyPath, err)
		}

		// delete sub keys
		if err := deleteSubKeysRecursively(rootKey, k, keyPath, access); err != nil {
			return err
		}

		if err := registry.DeleteKey(k, ""); err != nil {
			return fmt.Errorf("cannot delete key path : %q error: %v", keyPath, err)
		}
	}

	return nil
}

func deleteSubKeysRecursively(rootKey registry.Key, k registry.Key, rpath string, access uint32) error {
	subKeyNames, err := k.ReadSubKeyNames(0)
	if err != nil {
		return nil
	}

	for _, subKeyName := range subKeyNames {
		keyPath := fmt.Sprintf("%s\\%s", rpath, subKeyName)

		subRegKey, err := registry.OpenKey(rootKey, keyPath, access)
		if err != nil {
			return fmt.Errorf("path %q not found on registry: %v", subKeyName, err)
		}

		if err = deleteSubKeysRecursively(rootKey, subRegKey, keyPath, access); err != nil {
			return err
		}

		if err = registry.DeleteKey(subRegKey, ""); err != nil {
			return fmt.Errorf("cannot delete registry key path : %q error: %v", keyPath, err)
		}
	}

	return nil
}
