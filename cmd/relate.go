package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var relateCmd = &cobra.Command{
	Use:   "relate <ref> <target-id> [type]",
	Short: "Add a relation between two tasks",
	Args:  cobra.RangeArgs(2, 3),
	Example: `  grit relate auth 1776481293919
  grit relate auth 1776481293919 blocks
  grit relate auth 1776481293919 blocked-by`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := FindByRef(args[0])
		if err != nil {
			return err
		}
		targetID, err := strconv.ParseUint(strings.TrimSpace(args[1]), 10, 64)
		if err != nil {
			return fmt.Errorf("target id must be numeric, got %q", args[1])
		}
		relType := RelRelated
		if len(args) == 3 {
			relType = parseRelationType(args[2])
		}

		t, err := ParseTask(path)
		if err != nil {
			return err
		}
		found := false
		for i, r := range t.Relations {
			if r.TargetID == targetID {
				t.Relations[i].Type = relType
				found = true
				break
			}
		}
		if !found {
			t.Relations = append(t.Relations, Relation{TargetID: targetID, Type: relType})
		}
		if err := rewriteFrontmatter(path, t); err != nil {
			return err
		}

		fmt.Printf("grit: updated relations for task %d\n", t.ID)
		targetPath, err := FindByRef(fmt.Sprintf("%d", targetID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "grit: warning: target %d not found, inverse not written\n", targetID)
			return nil
		}
		return addInverseRelation(targetPath, t.ID, relType)
	},
}

var unrelateCmd = &cobra.Command{
	Use:   "unrelate <ref> <target-id>",
	Short: "Remove a relation between two tasks",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := FindByRef(args[0])
		if err != nil {
			return err
		}
		targetID, err := strconv.ParseUint(strings.TrimSpace(args[1]), 10, 64)
		if err != nil {
			return fmt.Errorf("target id must be numeric, got %q", args[1])
		}

		t, err := ParseTask(path)
		if err != nil {
			return err
		}
		filtered := t.Relations[:0]
		for _, r := range t.Relations {
			if r.TargetID != targetID {
				filtered = append(filtered, r)
			}
		}
		t.Relations = filtered
		if err := rewriteFrontmatter(path, t); err != nil {
			return err
		}

		targetPath, err := FindByRef(fmt.Sprintf("%d", targetID))
		if err != nil {
			return nil
		}
		return removeInverseRelation(targetPath, t.ID)
	},
}

func rewriteFrontmatter(path string, t Task) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	body := string(raw)
	rest := strings.TrimPrefix(body, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return fmt.Errorf("malformed frontmatter in %s", path)
	}
	bodyAfter := rest[idx+5:]

	newFront := fmt.Sprintf(
		"---\nid: %d\nstatus: %s\ntags: %s\nrelations: %s\n---\n",
		t.ID, t.Status,
		"["+joinTags(t.Tags)+"]",
		formatRelations(t.Relations),
	)

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(newFront+bodyAfter), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	return nil
}

func inverseType(r RelationType) RelationType {
	switch r {
	case RelParent:
		return RelChild
	case RelChild:
		return RelParent
	case RelBlocks:
		return RelBlockedBy
	case RelBlockedBy:
		return RelBlocks
	default:
		return RelRelated
	}
}

func addInverseRelation(targetPath string, sourceID uint64, relType RelationType) error {
	t, err := ParseTask(targetPath)
	if err != nil {
		return err
	}

	inv := inverseType(relType)

	for i, r := range t.Relations {
		if r.TargetID == sourceID {
			t.Relations[i].Type = inv
			return rewriteFrontmatter(targetPath, t)
		}
	}

	t.Relations = append(t.Relations, Relation{TargetID: sourceID, Type: inv})
	return rewriteFrontmatter(targetPath, t)
}

func removeInverseRelation(targetPath string, sourceID uint64) error {
	t, err := ParseTask(targetPath)
	if err != nil {
		return err
	}
	filtered := t.Relations[:0]
	for _, r := range t.Relations {
		if r.TargetID != sourceID {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == len(t.Relations) {
		return nil
	}
	t.Relations = filtered
	return rewriteFrontmatter(targetPath, t)
}

func init() {
	rootCmd.AddCommand(unrelateCmd)
	rootCmd.AddCommand(relateCmd)
}
