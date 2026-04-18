package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var addTags []string

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		id := uint64(time.Now().UnixNano() / 1e6)

		titleSlug := slug.Make(title)
		tags := "[]"
		if len(addTags) > 0 {
			tags = "[" + joinTags(addTags) + "]"
		}

		content := fmt.Sprintf("---\nid: %d\nstatus: open\ntags: %s\n---\n# %s\n", id, tags, title)
		filename := fmt.Sprintf("%d-%s.md", id, titleSlug)
		path := filepath.Join(TaskDir, filename)
		tmp := path + ".tmp"

		if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}

		fmt.Printf("grit: created %d - %s\n  -> %s\n", id, title, filename)
		return nil
	},
}

func init() {
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", nil, "comma-separated tags")
	rootCmd.AddCommand(addCmd)
}

func joinTags(tags []string) string {
	out := ""
	for i, t := range tags {
		if i > 0 {
			out += ", "
		}
		out += t
	}
	return out
}
