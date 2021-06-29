package cmd

import (
	"github.com/cloudradar-monitoring/tacoscript/script"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exeCmd)
}

const DefaultPath = "config.yaml"

var exeCmd = &cobra.Command{
	Use:   "exec {{PATH_TO_SCRIPT}}",
	Short: "Executes a script provided in argument, you can also run tacoscript {{PATH_TO_SCRIPT}}",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			args = []string{DefaultPath}
		}

		logrus.Debugf("will execute script %s", args[0])

		return script.RunScript(args[0])
	},
}
