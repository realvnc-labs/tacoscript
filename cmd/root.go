package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tacoscript",
		Short: "Tacoscript is a state-driven scripted task executor",
		Long:  "Tacoscript is a state-driven scripted task executor, See https://docs.saltstack.com/en/latest which is used for project inspiration",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("requires a PATH_TO_CMD_FILE argument")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			exeCmd.Run(cmd, args)
		},
		Version: Version,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
