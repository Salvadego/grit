package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Salvadego/qlvm"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	listQuery string
	listSort  string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks. Supports qlvm filter expressions.",
	Example: `  grit list
  grit list --query "status = 'open'"
  grit list -q ".bug AND @open"
  grit list -q "title ~ 'auth'"
  grit list -q "NOT @closed"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := filepath.Glob(TaskDir + "/*.md")
		if err != nil || len(files) == 0 {
			fmt.Println("grit: no tasks. run `grit add`")
			return nil
		}

		if listQuery == "" {
			listQuery = "@open" // show only open tasks by default
		}

		query := listQuery
		if len(args) > 0 {
			query = args[0]
		}

		cq, err := gritEngine.Compile(query)
		if err != nil {
			return fmt.Errorf("invalid query: %w", err)
		}

		cache := LoadState()
		dirty := false
		var tasks []Task

		for _, f := range files {
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			mtime := info.ModTime().UnixNano()
			entry, found := findEntryByFilename(cache, f)

			if found && entry.MTime == mtime {
				tasks = append(tasks, entry.ToTask(f))
				continue
			}

			t, err := ParseTask(f)
			if err != nil {
				continue
			}
			cache[t.ID], err = EntryFromTask(t, mtime)
			if err != nil {
				return err
			}

			dirty = true
			tasks = append(tasks, t)
		}

		if dirty {
			_ = SaveState(cache)
		}

		out, err := qlvm.FilterCompiled(cq, tasks, taskResolver)
		if err != nil {
			return err
		}

		if len(out) == 0 {
			fmt.Println("grit: no matching tasks.")
			return nil
		}

		sortTasks(out, listSort)
		renderTable(out)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listQuery, "query", "q", "", `filter expression, e.g. ".bug AND @open"`)
	listCmd.Flags().MarkHidden("query")

	listCmd.Flags().StringVarP(&listSort, "sort", "s", "created", "sort field: created, modified (append :asc for ascending)")
	listCmd.RegisterFlagCompletionFunc("sort", sortCompletionFn)

	rootCmd.AddCommand(listCmd)
}

func sortTasks(tasks []Task, by string) {
	asc := strings.HasSuffix(by, ":asc")
	field := strings.TrimSuffix(by, ":asc")

	sort.Slice(tasks, func(i, j int) bool {
		var a, b int64
		switch field {
		case "modified":
			a, b = tasks[i].MTime, tasks[j].MTime
		default:
			a, b = int64(tasks[i].ID), int64(tasks[j].ID)
		}
		if asc {
			return a < b
		}
		return a > b
	})
}

func findEntryByFilename(cache map[uint64]StateEntry, path string) (StateEntry, bool) {
	base := filepath.Base(path)
	var id uint64
	_, err := fmt.Sscanf(base, "%d-", &id)
	if err != nil {
		return StateEntry{}, false
	}
	e, ok := cache[id]
	return e, ok
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

func renderTable(tasks []Task) {
	const reset = "\033[0m"
	const green = "\033[32m"
	const gray = "\033[90m"
	const bold = "\033[1m"

	data := [][]string{}
	header := []string{
		"id",
		"status",
		"title",
		"tags",
	}

	for _, t := range tasks {
		var statusCol string
		if t.Status == "open" {
			statusCol = green + "open" + reset
		} else {
			statusCol = gray + "closed" + reset
		}

		title := t.Title
		if len(title) > 34 {
			title = title[:31] + "..."
		}
		tags := strings.Join(t.Tags, ", ")

		id := strconv.FormatUint(uint64(t.ID), 10)

		data = append(data, []string{
			id,
			statusCol,
			title,
			tags,
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(header)
	table.Bulk(data)
	table.Render()
}
