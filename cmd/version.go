package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.0.1"
	BuildTime = ""
	GitCommit = ""
	GitRef    = ""
)

func init() {
	rootCmd.AddCommand(versionCmd)
	BuildTime = time.Now().String()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Tacoscript",
	Long:  `All software has versions. This is Tacoscript's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`Version: %s
Build time: %s
Git commit: %s
Git ref: %s
`, Version, BuildTime, GitCommit, GitRef)
	},
}
