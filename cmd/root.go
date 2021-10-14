package cmd

import (
	"os"

	"github.com/cloudradar-monitoring/tacoscript/applog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Verbose      = false
	AbortOnError = false

	rootCmd = &cobra.Command{
		Use:           "taco",
		Short:         "Tacoscript is a state-driven scripted task executor",
		Args:          cobra.MinimumNArgs(1),
		RunE:          exeCmd.RunE,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version(),
	}
)

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&AbortOnError, "abort-on-error", "a", false, "Abort on error")
}

func initLog() {
	applog.Init(Verbose)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		logrus.Debugf("Execute failed: %v", err)
		os.Exit(1)
	}

	return nil
}
