package cmd

import (
	"fmt"
)

var (
	Version   = ""
	BuildTime = ""
	GitCommit = ""
)

func version() string {
	return fmt.Sprintf(`Version: %s
Build time: %s
Git commit: %s
`, Version, BuildTime, GitCommit)
}
