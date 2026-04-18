package cmd

import "strings"

const TaskDir = ".grit"
const StateFile = ".grit/.state.bin"
const StateMagic uint32 = 0x47726954
const StateVersion uint8 = 3

type RelationType uint8

const (
	RelRelated RelationType = iota
	RelParent
	RelChild
	RelBlocks
	RelBlockedBy
)

func (r RelationType) String() string {
	switch r {
	case RelParent:
		return "parent"
	case RelChild:
		return "child"
	case RelBlocks:
		return "blocks"
	case RelBlockedBy:
		return "blocked-by"
	default:
		return "related"
	}
}

func parseRelationType(s string) RelationType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "parent":
		return RelParent
	case "child":
		return RelChild
	case "blocks":
		return RelBlocks
	case "blocked-by", "blockedby":
		return RelBlockedBy
	default:
		return RelRelated
	}
}

type Relation struct {
	TargetID uint64
	Type     RelationType
}

type Task struct {
	ID        uint64
	Status    string
	MTime     int64
	Title     string
	Tags      []string
	Relations []Relation
	Filename  string
}
