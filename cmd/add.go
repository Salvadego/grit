package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var addTags []string
var addRelations []string

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

		var rels []Relation
		for _, raw := range addRelations {
			idStr, relType, hasType := strings.Cut(raw, ":")
			id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid relation %q: id must be numeric", raw)
			}
			rel := Relation{TargetID: id}
			if hasType {
				rel.Type = parseRelationType(relType)
			}
			rels = append(rels, rel)
		}

		content := fmt.Sprintf(
			"---\nid: %d\nstatus: open\ntags: %s\nrelations: %s\n---\n# %s\n",
			id, tags, formatRelations(rels), title,
		)
		filename := fmt.Sprintf("%d-%s.md", id, titleSlug)
		path := filepath.Join(TaskDir, filename)
		tmp := path + ".tmp"

		if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}

		for _, rel := range rels {
			targetPath, err := FindByRef(fmt.Sprintf("%d", rel.TargetID))
			if err != nil {
				fmt.Fprintf(os.Stderr, "grit: warning: relation target %d not found, skipping inverse\n", rel.TargetID)
				continue
			}
			if err := addInverseRelation(targetPath, id, rel.Type); err != nil {
				fmt.Fprintf(os.Stderr, "grit: warning: could not write inverse relation to %d: %v\n", rel.TargetID, err)
			}
		}

		fmt.Printf("grit: created %d - %s\n  -> %s\n", id, title, filename)
		return nil
	},
}

func init() {
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", nil, "comma-separated tags")
	addCmd.Flags().StringSliceVarP(&addRelations, "relations", "r", nil, `related task IDs, e.g. -r 1776481293919 -r 1776481439085:blocks`)
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
