// +build windows

package apptest

func AssertFileMatchesExpectationOS(filePath string, fe *FileExpectation) (bool, string, error) {
	return true, "", nil
}
