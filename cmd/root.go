package cmd

import (
	"github.com/cloudradar-monitoring/tacoscript/applog"
	"github.com/spf13/cobra"
)

var (
	Verbose      = false
	AbortOnError = false

	rootCmd = &cobra.Command{
		Use:          "taco",
		Short:        "Tacoscript is a state-driven scripted task executor",
		Args:         cobra.MinimumNArgs(1),
		RunE:         exeCmd.RunE,
		SilenceUsage: true,
		Version:      version(),
	}
)

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&AbortOnError, "abort-on-error", "", false, "Abort on error")
}

func initLog() {
	applog.Init(Verbose)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
