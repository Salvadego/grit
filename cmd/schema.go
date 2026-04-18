package cmd

import (
	"fmt"
	"strings"

	"github.com/Salvadego/qlvm"
)

var gritEngine = qlvm.New(
	qlvm.NewSchema().
		Field("title", qlvm.String).
		Field("status", qlvm.String).
		Field("tags", qlvm.String).
		Field("id", qlvm.Number).
		Field("relations", qlvm.String).
		Field("blocked", qlvm.Bool).
		Field("blocks", qlvm.String).
		Field("blocked-by", qlvm.String).
		Field("parent", qlvm.String).
		Field("child", qlvm.String).
		Field("related", qlvm.String).
		Prefix('.', "tags", qlvm.Contains).
		Prefix('@', "status", qlvm.EQ).
		Prefix('!', "relations", qlvm.Contains).
		Suffix('?', qlvm.Exists()),
)

func taskResolver(t Task) qlvm.Resolver {
	buckets := map[RelationType][]string{
		RelRelated:   {},
		RelParent:    {},
		RelChild:     {},
		RelBlocks:    {},
		RelBlockedBy: {},
	}
	allParts := make([]string, 0, len(t.Relations))
	isBlocked := false

	for _, r := range t.Relations {
		idStr := fmt.Sprintf("%d", r.TargetID)
		buckets[r.Type] = append(buckets[r.Type], idStr)
		allParts = append(allParts, idStr+":"+r.Type.String())
		if r.Type == RelBlockedBy {
			isBlocked = true
		}
	}

	join := func(ids []string) string { return strings.Join(ids, " ") }

	return func(field string) (any, bool) {
		switch field {
		case "title":
			return t.Title, true
		case "status":
			return t.Status, true
		case "tags":
			return strings.Join(t.Tags, " "), true
		case "id":
			return float64(t.ID), true
		case "relations":
			return strings.Join(allParts, " "), true
		case "blocked":
			return isBlocked, true
		case "blocks":
			return join(buckets[RelBlocks]), true
		case "blocked-by":
			return join(buckets[RelBlockedBy]), true
		case "parent":
			return join(buckets[RelParent]), true
		case "child":
			return join(buckets[RelChild]), true
		case "related":
			return join(buckets[RelRelated]), true
		}
		return nil, false
	}
}
