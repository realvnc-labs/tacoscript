package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = "v0.0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Tacoscript",
	Long:  `All software has versions. This is Tacoscript's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
