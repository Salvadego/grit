package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func refCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cache := LoadState()
	var suggestions []string

	for _, entry := range cache {
		// we keep them all for now, cases like the close cmd and reopen command
		// if entry.StatusByte != 0 {
		// 	continue
		// }

		slug := nullTrimmed(entry.Slug[:])
		title := nullTrimmed(entry.Title[:])
		id := fmt.Sprintf("%d", entry.ID)

		if toComplete == "" || strings.HasPrefix(slug, toComplete) {
			suggestions = append(suggestions, slug+"\t"+title)
		}

		if toComplete != "" && strings.HasPrefix(id, toComplete) {
			suggestions = append(suggestions, id+"\t"+title)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	refCommands := []*cobra.Command{
		editCmd,
		closeCmd,
		reopenCmd,
		deleteCmd,
		relateCmd,
		unrelateCmd,
	}
	for _, cmd := range refCommands {
		cmd.ValidArgsFunction = refCompletionFn
	}
}
