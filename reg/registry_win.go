//go:build windows
// +build windows

package reg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func GetValue(regPath string, name string, valType RegistryType) (found bool, val any, err error) {
	key, keyPath, err := getRootKey(regPath)
	if err != nil {
		return false, val, err
	}

	k, err := registry.OpenKey(key, keyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// return not found with no error
			return false, val, nil
		}
		return false, val, err
	}
	defer k.Close()

	val, _, err = getValueByType(k, name, valType)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// return not found with no error
			return false, val, nil
		}
		// unexpected type means the value was still found, but with an error
		if errors.Is(err, registry.ErrUnexpectedType) {
			return true, val, err
		}
		return false, val, err
	}
	return true, val, nil
}

func SetValue(regPath string, name string, val any, valType RegistryType) (updated bool, desc string, err error) {
	key, keyPath, err := getRootKey(regPath)
	if err != nil {
		return false, "", err
	}

	k, keyExists, err := registry.CreateKey(key, keyPath, registry.ALL_ACCESS)
	if err != nil {
		fmt.Printf("err = %+v\n", err)
		return false, "", ErrFailedCreatingOrOpeningKey
	}
	defer k.Close()

	// handle when new key
	if !keyExists {
		err = setValueByType(k, name, val, valType)
		if err != nil {
			return false, "", err
		}
		return true, "added new key", nil
	}

	// handle when existing key
	valExists := true
	existingVal, actualType, err := getValueByType(k, name, valType)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			valExists = false
		} else {
			if errors.Is(err, registry.ErrUnexpectedType) {
				valExists = true
			} else {
				return false, "", err
			}
		}
	}

	rType, err := getRegistryType(valType)
	if err != nil {
		return false, "", err
	}

	if valExists {
		if actualType == rType {
			match, err := compareValueByType(val, existingVal, valType)
			if err != nil {
				return false, "", err
			}
			if match {
				return false, "matching existing value", nil
			}
		} else {
			// remove existing value with the wrong type
			err := k.DeleteValue(name)
			if err != nil {
				return false, "", err
			}
		}
	}

	err = setValueByType(k, name, val, valType)
	if err != nil {
		return false, "", err
	}

	if valExists {
		return true, "existing value updated", nil
	}

	return true, "added new value", nil
}

func RemoveValue(regPath string, name string) (updated bool, desc string, err error) {
	key, keyPath, err := getRootKey(regPath)
	if err != nil {
		return false, "", err
	}

	found, _, err := GetValue(regPath, name, REG_SZ)
	if err != nil && !found {
		return false, "", err
	}

	if !found {
		return false, "no existing value", nil
	}

	k, err := registry.OpenKey(key, keyPath, registry.ALL_ACCESS)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, "no existing key", nil
		}
		return false, "", err
	}
	defer k.Close()

	err = k.DeleteValue(name)
	if err != nil {
		return false, "", err
	}

	return true, "value removed", nil
}

func RemoveKey(regPath string) (updated bool, desc string, err error) {
	err = DeleteKeyRecursive(regPath)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// key not found is not error. return nothing changed
			return false, "no existing key", nil
		}
		return false, "", err
	}

	// key was deleted
	return true, "key removed", nil
}

func HasValidRootKey(rootKey string) (err error) {
	_, _, err = getRootKey(rootKey)
	return err
}

func getValueByType(k registry.Key, name string, valType RegistryType) (val any, actualType uint32, err error) {
	var intVal uint64

	switch valType {
	case REG_SZ:
		val, actualType, err = k.GetStringValue(name)
		return val, actualType, err
	case REG_BINARY:
		val, actualType, err = k.GetBinaryValue(name)
		return val, actualType, err
	case REG_DWORD:
		intVal, actualType, err = k.GetIntegerValue(name)
		return uint32(intVal), actualType, err
	case REG_QWORD:
		intVal, actualType, err = k.GetIntegerValue(name)
		return intVal, actualType, err
	default:
		return nil, 0, ErrUnknownValType
	}
}

func setValueByType(k registry.Key, name string, val any, valType RegistryType) (err error) {
	switch valType {
	case REG_SZ:
		strVal, ok := val.(string)
		if !ok {
			return ErrFailedToConvertVal
		}
		err = k.SetStringValue(name, strVal)
	case REG_BINARY:
		binVal, ok := val.([]byte)
		if !ok {
			return ErrFailedToConvertVal
		}
		err = k.SetBinaryValue(name, binVal)
	case REG_DWORD:
		intVal, ok := val.(uint32)
		if !ok {
			return ErrFailedToConvertVal
		}
		err = k.SetDWordValue(name, intVal)
	case REG_QWORD:
		intVal, ok := val.(uint64)
		if !ok {
			return ErrFailedToConvertVal
		}
		err = k.SetQWordValue(name, intVal)
	default:
		err = ErrUnknownValType
	}

	return err
}

func compareValueByType(val1 any, val2 any, valType RegistryType) (match bool, err error) {
	switch valType {
	case REG_SZ:
		strVal1, ok := val1.(string)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		strVal2, ok := val2.(string)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		match = strings.Compare(strVal1, strVal2) == 0
	case REG_BINARY:
		binVal1, ok := val1.([]byte)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		binVal2, ok := val2.([]byte)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		match = bytes.Equal(binVal1, binVal2)
	case REG_DWORD:
		intVal1, ok := val1.(uint32)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		intVal2, ok := val2.(uint32)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		match = intVal1 == intVal2
	case REG_QWORD:
		intVal1, ok := val1.(uint64)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		intVal2, ok := val2.(uint64)
		if !ok {
			return false, ErrFailedToConvertVal
		}
		match = intVal1 == intVal2
	default:
		err = ErrUnknownValType
	}

	return match, err
}

func getRegistryType(valType RegistryType) (regType uint32, err error) {
	switch valType {
	case REG_SZ:
		regType = registry.SZ
	case REG_BINARY:
		regType = registry.BINARY
	case REG_DWORD:
		regType = registry.DWORD
	case REG_QWORD:
		regType = registry.QWORD
	default:
		err = ErrUnknownValType
		return 0, err
	}
	return regType, nil
}

func getRootKey(regPath string) (rootKey registry.Key, keyPath string, err error) {
	rootKeyStr, keyPath, found := strings.Cut(regPath, `:\`)
	if err != nil {
		return 0, "", err
	}
	if !found {
		return 0, "", ErrMissingRootKey
	}
	rootKey, err = mapRootKeyToRegistryKey(rootKeyStr)
	if err != nil {
		return 0, "", err
	}
	return rootKey, keyPath, nil
}

func mapRootKeyToRegistryKey(rootKey string) (key registry.Key, err error) {
	switch rootKey {
	case "HKCU":
		return registry.CURRENT_USER, nil
	case "HKLM":
		return registry.LOCAL_MACHINE, nil
	case "HKU":
		return registry.USERS, nil
	default:
		return 0, ErrUnknownRootKey
	}
}
