package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <id|slug>",
	Short: "Mark a task as closed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setStatus(args[0], "open", "closed")
	},
}

func setStatus(ref, from, to string) error {
	path, err := FindByRef(ref)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	body := string(raw)
	if !strings.Contains(body, "status: "+from) {
		return fmt.Errorf("task is already %s", to)
	}
	body = strings.Replace(body, "status: "+from, "status: "+to, 1)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(body), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	id, _ := idFromFilename(path)
	if isPiped() {
		fmt.Printf("%d\n", id)
	} else {
		fmt.Printf("grit: %s -> %s\n", ref, to)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(closeCmd)
}
