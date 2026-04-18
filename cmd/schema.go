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
		Prefix('.', "tags", qlvm.Contains).
		Prefix('@', "status", qlvm.EQ).
		Prefix('!', "relations", qlvm.Contains).
		Suffix('?', qlvm.Exists()),
)

func taskResolver(t Task) qlvm.Resolver {
	relParts := make([]string, len(t.Relations))
	for i, r := range t.Relations {
		relParts[i] = fmt.Sprintf("%d:%s", r.TargetID, r.Type)
	}
	relStr := strings.Join(relParts, " ")

	isBlocked := false
	for _, r := range t.Relations {
		if r.Type == RelBlockedBy {
			isBlocked = true
			break
		}
	}

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
			return relStr, true
		case "blocked":
			return isBlocked, true
		}
		return nil, false
	}
}
