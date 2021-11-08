//go:build windows
// +build windows

package apptest

func AssertFileMatchesExpectationOS(filePath string, fe *FileExpectation) (isMatched bool, reason string, err error) {
	return true, "", nil
}
