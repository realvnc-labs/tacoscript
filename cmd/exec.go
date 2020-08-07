package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exeCmd)
}

var exeCmd = &cobra.Command{
	Use:   "exec {{PATH_TO_CMD_FILE}}",
	Short: "Executes script provided in argument, you can also run tacoscript {{PATH_TO_CMD_FILE}}",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a PATH_TO_CMD_FILE argument")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("EXEC", args)
	},
}
