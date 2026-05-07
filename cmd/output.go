package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

func isPiped() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

type column struct {
	header   string
	value    func(t Task) string
	minWidth int
}

var taskColumns = []column{
	{
		header:   "id",
		minWidth: 0,
		value:    func(t Task) string { return fmt.Sprintf("%d", t.ID) },
	},
	{
		header:   "status",
		minWidth: 0,
		value: func(t Task) string {
			if t.Status == "open" {
				return "\033[32mopen\033[0m"
			}
			return "\033[90m" + t.Status + "\033[0m"
		},
	},
	{
		header:   "title",
		minWidth: 0,
		value:    func(t Task) string { return t.Title },
	},
	{
		header:   "tags",
		minWidth: 100,
		value:    func(t Task) string { return strings.Join(t.Tags, ", ") },
	},
	{
		header:   "relations",
		minWidth: 130,
		value: func(t Task) string {
			if len(t.Relations) == 0 {
				return ""
			}
			if len(t.Relations) > 3 {
				counts := make(map[string]int)
				for _, r := range t.Relations {
					counts[r.Type.String()]++
				}
				parts := make([]string, 0, len(counts))
				for typ, n := range counts {
					parts = append(parts, fmt.Sprintf("%d %s", n, typ))
				}
				return fmt.Sprintf("[%d] %s", len(t.Relations), strings.Join(parts, ", "))
			}
			parts := make([]string, len(t.Relations))
			for i, r := range t.Relations {
				id := fmt.Sprintf("%d", r.TargetID)
				if len(id) > 6 {
					id = "…" + id[len(id)-6:]
				}
				parts[i] = r.Type.String() + ":" + id
			}
			return strings.Join(parts, ", ")
		},
	},
	{
		header:   "created",
		minWidth: 150,
		value:    func(t Task) string { return time.UnixMilli(int64(t.ID)).Format("2006-01-02") },
	},
	{
		header:   "modified",
		minWidth: 170,
		value:    func(t Task) string { return time.Unix(0, t.MTime).Format("2006-01-02") },
	},
}

func visibleColumns(width int) []column {
	var out []column
	for _, c := range taskColumns {
		if width >= c.minWidth {
			out = append(out, c)
		}
	}
	return out
}

func titleMaxWidth(termW, numCols int) int {
	reserved := (numCols - 1) * 18
	max := termW - reserved - 4
	if max < 20 {
		return 20
	}
	if max > 60 {
		return 60
	}
	return max
}

func renderTable(tasks []Task) {
	width := termWidth()
	cols := visibleColumns(width)
	titleMax := titleMaxWidth(width, len(cols))

	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = c.header
	}

	data := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		row := make([]string, len(cols))
		for i, c := range cols {
			v := c.value(t)
			if c.header == "title" && len(v) > titleMax {
				v = v[:titleMax-3] + "..."
			}
			row[i] = v
		}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)
	table.Bulk(data)
	table.Render()
}

func renderPiped(tasks []Task) {
	fmt.Println("id\tstatus\ttitle\ttags\trelations\tcreated\tmodified")
	for _, t := range tasks {
		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			t.Status,
			t.Title,
			strings.Join(t.Tags, ","),
			formatRelationsTSV(t.Relations),
			time.UnixMilli(int64(t.ID)).Format("2006-01-02"),
			time.Unix(0, t.MTime).Format("2006-01-02"),
		)
	}
}

func formatRelationsShort(rels []Relation) string {
	if len(rels) == 0 {
		return ""
	}
	parts := make([]string, len(rels))
	for i, r := range rels {
		parts[i] = r.Type.String() + ":" + fmt.Sprintf("%d", r.TargetID)
	}
	return strings.Join(parts, ", ")
}

func formatRelationsTSV(rels []Relation) string {
	if len(rels) == 0 {
		return ""
	}
	parts := make([]string, len(rels))
	for i, r := range rels {
		parts[i] = r.Type.String() + ":" + fmt.Sprintf("%d", r.TargetID)
	}
	return strings.Join(parts, "|")
}
