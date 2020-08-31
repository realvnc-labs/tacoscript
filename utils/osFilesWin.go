// +build windows

package utils

func AssertFileMatchesExpectationOS(filePath string, fe *FileExpectation) (bool, string, error) {
	return true, "", nil
}

func ParseLocationOS(rawLocation string) string {
	if !strings.HasPrefix(rawLocation, "file:") {
		return rawLocation
	}

	rawLocation = strings.TrimPrefix(rawLocation, "file:///")
	rawLocation = strings.Replace(rawLocation, "/", string(os.PathSeparator), -1)

	return rawLocation
}
