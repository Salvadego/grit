package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <id|slug>",
	Aliases: []string{"rm"},
	Short:   "Permanently delete a task file",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := FindByRef(args[0])
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		fmt.Printf("grit: deleted %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
