package cmd

import (
	"strings"

	"github.com/Salvadego/qlvm"
)

var gritEngine = qlvm.New(
	qlvm.NewSchema().
		Field("title", qlvm.String).
		Field("status", qlvm.String).
		Field("tags", qlvm.String).
		Field("id", qlvm.Number).
		Prefix('.', "tags", qlvm.Contains). // .bug       -> tags CONTAINS 'bug'
		Prefix('@', "status", qlvm.EQ).     // @open      -> status = 'open'
		Suffix('?', qlvm.Exists()),         // title?     -> title != ''
)

func taskResolver(t Task) qlvm.Resolver {
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
		}
		return nil, false
	}
}
