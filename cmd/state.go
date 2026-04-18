package cmd

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type StateHeader struct {
	Magic   uint32
	Version uint8
	Count   uint64
}

type RelationEntry struct {
	TargetID uint64
	RelType  uint8
}

type StateEntry struct {
	ID            uint64
	MTime         int64
	StatusByte    uint8
	TagCount      uint8
	Title         [256]byte
	Slug          [128]byte
	Tags          [8][64]byte
	RelationCount uint8
	Relations     [8]RelationEntry
}

const maxTags = 8
const maxRelations = 8
const entrySize = 8 + 8 + 1 + 1 + 256 + 128 + 8*64

func (e StateEntry) ToTask(filename string) Task {
	rels := make([]Relation, 0, e.RelationCount)
	for i := uint8(0); i < e.RelationCount; i++ {
		rels = append(rels, Relation{
			TargetID: e.Relations[i].TargetID,
			Type:     RelationType(e.Relations[i].RelType),
		})
	}
	return Task{
		ID:        e.ID,
		Status:    statusString(e.StatusByte),
		MTime:     e.MTime,
		Title:     nullTrimmed(e.Title[:]),
		Tags:      decodeTags(e),
		Relations: rels,
		Filename:  filename,
	}
}

func EntryFromTask(t Task, mtime int64) StateEntry {
	var e StateEntry
	e.ID = t.ID
	e.MTime = mtime
	e.StatusByte = statusByte(t.Status)
	copyFixed(e.Title[:], t.Title)
	copyFixed(e.Slug[:], slugFromFilename(t.Filename))

	n := len(t.Tags)
	if n > maxTags {
		n = maxTags
	}
	e.TagCount = uint8(n)
	for i := 0; i < n; i++ {
		copyFixed(e.Tags[i][:], t.Tags[i])
	}

	rn := len(t.Relations)
	if rn > maxRelations {
		rn = maxRelations
	}
	e.RelationCount = uint8(rn)
	for i := 0; i < rn; i++ {
		e.Relations[i] = RelationEntry{
			TargetID: t.Relations[i].TargetID,
			RelType:  uint8(t.Relations[i].Type),
		}
	}
	return e
}

func LoadState() map[uint64]StateEntry {
	out := make(map[uint64]StateEntry)

	f, err := os.Open(StateFile)
	if err != nil {
		return out
	}
	defer f.Close()

	var hdr StateHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return out
	}
	if hdr.Magic != StateMagic || hdr.Version != StateVersion {
		return out
	}

	var i uint64
	for i = 0; i < hdr.Count; i++ {
		var e StateEntry
		if err := binary.Read(f, binary.LittleEndian, &e); err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		out[e.ID] = e
	}
	return out
}

func SaveState(entries map[uint64]StateEntry) error {
	tmp := StateFile + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	hdr := StateHeader{
		Magic:   StateMagic,
		Version: StateVersion,
		Count:   uint64(len(entries)),
	}
	if err := binary.Write(f, binary.LittleEndian, hdr); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	for _, e := range entries {
		if err := binary.Write(f, binary.LittleEndian, e); err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
	}

	f.Close()
	return os.Rename(tmp, StateFile)
}

func RebuildState() (map[uint64]StateEntry, error) {
	files, _ := filepath.Glob(TaskDir + "/*.md")
	entries := make(map[uint64]StateEntry, len(files))

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		t, err := ParseTask(f)
		if err != nil {
			continue
		}
		entries[t.ID] = EntryFromTask(t, info.ModTime().UnixNano())
	}

	return entries, SaveState(entries)
}

func statusByte(s string) uint8 {
	if strings.TrimSpace(s) == "closed" {
		return 1
	}
	return 0
}

func statusString(b uint8) string {
	if b == 1 {
		return "closed"
	}
	return "open"
}

func copyFixed(dst []byte, s string) {
	n := len(s)
	if n > len(dst) {
		n = len(dst)
	}
	copy(dst[:n], s[:n])
}

func nullTrimmed(b []byte) string {
	s := string(b)
	if idx := strings.IndexByte(s, 0); idx >= 0 {
		return s[:idx]
	}
	return s
}

func decodeTags(e StateEntry) []string {
	tags := make([]string, 0, e.TagCount)
	for i := uint8(0); i < e.TagCount; i++ {
		if t := nullTrimmed(e.Tags[i][:]); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func slugFromFilename(filename string) string {
	base := filepath.Base(filename)
	parts := strings.SplitN(base, "-", 2)
	if len(parts) == 2 {
		return strings.TrimSuffix(parts[1], ".md")
	}
	return base
}

func EntryMatchesRef(e StateEntry, ref string) bool {
	slug := nullTrimmed(e.Slug[:])
	idStr := fmt.Sprintf("%d", e.ID)
	// fmt.Println(idStr, ref)
	// fmt.Println(strings.HasPrefix(idStr, ref))
	return strings.HasPrefix(idStr, ref) || strings.EqualFold(slug, ref) || strings.Contains(slug, ref)
}
