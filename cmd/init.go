package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes grit",
	Long:  `Creates the .grit folder in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		initGrit()
	},
}

func initGrit() {
	os.Mkdir(TaskDir, 0755)
	fmt.Printf("Initialized Grit in %s\n", TaskDir)
}

func init() {
	rootCmd.AddCommand(initCmd)
}
