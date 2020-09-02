package utils

import (
	"fmt"

	"github.com/kylelemons/godebug/diff"
)

func Diff(expectedStr, actualStr string) string {
	contentDiff := diff.Diff(expectedStr, actualStr)
	if contentDiff == "" {
		return ""
	}

	return fmt.Sprintf(`expected: "%v"
actual: "%v"
Diff:
--- Expected
+++ Actual
%s
`, Truncate(expectedStr), Truncate(actualStr), contentDiff)
}
