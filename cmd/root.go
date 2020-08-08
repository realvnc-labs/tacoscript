package cmd

import (
	"github.com/cloudradar-monitoring/tacoscript/applog"
	"github.com/spf13/cobra"
)

var (
	Verbose = false
	rootCmd = &cobra.Command{
		Use:   "tacoscript",
		Short: "Tacoscript is a state-driven scripted task executor",
		Long:  "Tacoscript is a state-driven scripted task executor, See https://docs.saltstack.com/en/latest which is used for project inspiration",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exeCmd.RunE(cmd, args)
		},
		Version: Version,
	}
)

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
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
