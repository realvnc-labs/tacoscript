package cmd

import (
	"github.com/cloudradar-monitoring/tacoscript/script"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exeCmd)
}

var exeCmd = &cobra.Command{
	Use:   "exec [script to run]",
	Short: "Executes a script provided in argument, you can also run taco {{PATH_TO_SCRIPT}}",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Debugf("will execute script %s (abort-on-error=%v)", args[0], AbortOnError)

		return script.RunScript(args[0], AbortOnError)
	},
	SilenceErrors: true,
}
