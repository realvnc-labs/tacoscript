//go:build !windows
// +build !windows

package winregistry

import "errors"

var (
	ErrFnNotImplemented = errors.New("registry fns only supported on windows")
)

func GetValue(regPath string, name string, valType RegistryType) (found bool, val any, err error) {
	return false, nil, ErrFnNotImplemented
}

func SetValue(regPath string, name string, val any, valType RegistryType) (updated bool, desc string, err error) {
	return false, "", ErrFnNotImplemented
}

func RemoveValue(regPath string, name string) (updated bool, desc string, err error) {
	return false, "", ErrFnNotImplemented
}

func RemoveKey(regPath string) (updated bool, desc string, err error) {
	return false, "", ErrFnNotImplemented
}

func HasValidRootKey(rootKey string) (err error) {
	return ErrFnNotImplemented
}
