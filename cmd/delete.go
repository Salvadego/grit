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

		t, err := ParseTask(path)
		if err != nil {
			return err
		}

		if err := os.Remove(path); err != nil {
			return err
		}

		cache := LoadState()
		if _, exists := cache[t.ID]; exists {
			delete(cache, t.ID)
			if err := SaveState(cache); err != nil {
				fmt.Fprintf(os.Stderr, "grit: warning: deleted file but failed to update state: %v\n", err)
			}
		}

		fmt.Printf("grit: deleted %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
