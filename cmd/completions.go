package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func queryCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	lastTok, prefix := lastToken(toComplete)

	var suggestions []string

	switch {
	case strings.HasPrefix(lastTok, "@"):
		partial := strings.TrimPrefix(lastTok, "@")
		for _, v := range []string{"open", "closed"} {
			if strings.HasPrefix(v, partial) {
				suggestions = append(suggestions,
					prefix+"@"+v+"\t"+"status = "+v)
			}
		}

	case strings.HasPrefix(lastTok, "."):
		partial := strings.TrimPrefix(lastTok, ".")
		for _, tag := range liveTagSuggestions(partial) {
			suggestions = append(suggestions,
				prefix+"."+tag+"\ttags contains "+tag)
		}
	case strings.HasPrefix(lastTok, "!"):
		partial := strings.TrimPrefix(lastTok, "!")
		for _, s := range liveRefSuggestions(partial) {
			slug := strings.SplitN(s, "\t", 2)[0]
			title := ""
			if parts := strings.SplitN(s, "\t", 2); len(parts) == 2 {
				title = parts[1]
			}
			suggestions = append(suggestions, prefix+"!"+slug+"\t"+"relates to: "+title)
		}

	default:
		fields := []struct{ name, desc string }{
			{"title", "string - task title"},
			{"status", "string - open or closed"},
			{"tags", "string - space-joined tag list"},
			{"id", "number - task id"},
			{"relations", "string - all relations (id:type ...)"},
			{"blocked", "bool   - has any blocked-by relation"},
			{"blocks", "string - IDs of tasks this blocks"},
			{"blocked-by", "string - IDs of tasks blocking this"},
			{"parent", "string - IDs of parent tasks"},
			{"child", "string - IDs of child tasks"},
			{"related", "string - IDs of related tasks"},
			{"created", "date - derived from task ID"},
			{"modified", "date - last file modification"},
		}
		keywords := []struct{ name, desc string }{
			{"AND", "logical and"},
			{"OR", "logical or"},
			{"NOT", "logical not"},
		}

		for _, f := range fields {
			if strings.HasPrefix(f.name, strings.ToLower(lastTok)) {
				suggestions = append(suggestions, prefix+f.name+"\t"+f.desc)
			}
		}
		for _, k := range keywords {
			if strings.HasPrefix(k.name, strings.ToUpper(lastTok)) {
				suggestions = append(suggestions, prefix+k.name+"\t"+k.desc)
			}
		}
		if isRelationValueContext(toComplete) {
			for _, s := range liveRefSuggestions("") {
				suggestions = append(suggestions, prefix+s)
			}
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

func lastToken(s string) (lastTok, prefix string) {
	i := len(s)
	for i > 0 {
		ch := rune(s[i-1])
		if ch == ' ' || ch == '(' {
			break
		}
		i--
	}
	return s[i:], s[:i]
}

func isRelationValueContext(q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	return strings.HasSuffix(q, "relations contains") ||
		strings.HasSuffix(q, "blocks contains") ||
		strings.HasSuffix(q, "blocked-by contains") ||
		strings.HasSuffix(q, "parent contains") ||
		strings.HasSuffix(q, "child contains") ||
		strings.HasSuffix(q, "related contains") ||
		strings.HasSuffix(q, "relations =") ||
		strings.HasSuffix(q, "!")
}

func liveTagSuggestions(partial string) []string {
	cache := LoadState()
	seen := make(map[string]bool)
	var out []string
	for _, entry := range cache {
		if entry.StatusByte != 0 {
			continue
		}
		for i := uint8(0); i < entry.TagCount; i++ {
			tag := nullTrimmed(entry.Tags[i][:])
			if seen[tag] || !strings.HasPrefix(tag, partial) {
				continue
			}
			seen[tag] = true
			out = append(out, tag)
		}
	}
	return out
}

func liveRefSuggestions(partial string) []string {
	cache := LoadState()
	var out []string
	for _, entry := range cache {
		// if entry.StatusByte != 0 {
		// 	continue
		// }
		slug := nullTrimmed(entry.Slug[:])
		title := nullTrimmed(entry.Title[:])
		id := fmt.Sprintf("%d", entry.ID)
		if partial == "" || strings.HasPrefix(slug, partial) {
			out = append(out, slug+"\t"+title)
		}
		if partial != "" && strings.HasPrefix(id, partial) {
			out = append(out, id+"\t"+title)
		}
	}
	return out
}

func refCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return liveRefSuggestions(toComplete), cobra.ShellCompDirectiveNoFileComp
}

func sortCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	options := []struct{ val, desc string }{
		{"created", "newest created first (default)"},
		{"created:asc", "oldest created first"},
		{"created:desc", "newest created first"},
		{"modified", "newest modified first"},
		{"modified:asc", "oldest modified first"},
		{"modified:desc", "newest modified first"},
	}
	var out []string
	for _, o := range options {
		if strings.HasPrefix(o.val, toComplete) {
			out = append(out, o.val+"\t"+o.desc)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func execCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	for _, prefix := range []string{"tag:", "untag:"} {
		if strings.HasPrefix(toComplete, prefix) {
			partial := strings.TrimPrefix(toComplete, prefix)
			var out []string
			for _, tag := range liveTagSuggestions(partial) {
				out = append(out, prefix+tag+"\tapply to matched tasks")
			}
			return out, cobra.ShellCompDirectiveNoFileComp
		}
	}

	options := []struct{ val, desc string }{
		{"close", "mark all matched tasks as closed"},
		{"reopen", "reopen all matched tasks"},
		{"delete", "permanently delete all matched tasks"},
		{"tag:", "add a tag, e.g. tag:v0.3.0"},
		{"untag:", "remove a tag, e.g. untag:v0.2.0"},
	}
	var out []string
	for _, o := range options {
		if strings.HasPrefix(o.val, toComplete) {
			out = append(out, o.val+"\t"+o.desc)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
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

	listCmd.ValidArgsFunction = queryCompletionFn
}
