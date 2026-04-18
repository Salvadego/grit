package cmd

import "github.com/spf13/cobra"

var reopenCmd = &cobra.Command{
	Use:   "reopen <id|slug>",
	Short: "Reopen a closed task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setStatus(args[0], "closed", "open")
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
