package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseTask(path string) (Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return Task{}, err
	}
	defer f.Close()

	var t Task
	t.Filename = path

	scanner := bufio.NewScanner(f)
	inFront, frontDone := false, false
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if lineNum == 1 && line == "---" {
			inFront = true
			continue
		}
		if inFront && line == "---" {
			frontDone = true
			inFront = false
			continue
		}
		if inFront {
			key, val, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			switch key {
			case "id":
				n, _ := strconv.ParseUint(val, 10, 64)
				t.ID = n
			case "status":
				t.Status = val
			case "tags":
				val = strings.Trim(val, "[]")
				if val != "" {
					for _, tag := range strings.Split(val, ",") {
						t.Tags = append(t.Tags, strings.TrimSpace(tag))
					}
				}
			}
			continue
		}
		if frontDone && strings.HasPrefix(line, "# ") {
			t.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	if t.ID == 0 {
		return Task{}, fmt.Errorf("missing id in %s", path)
	}
	if t.Title == "" {
		t.Title = "(untitled)"
	}
	if t.Status == "" {
		t.Status = "open"
	}
	return t, nil
}

func FindByRef(ref string) (string, error) {
	cache := LoadState()

	for id, entry := range cache {
		if EntryMatchesRef(entry, ref) {
			slug := nullTrimmed(entry.Slug[:])
			filename := fmt.Sprintf("%d-%s.md", id, slug)
			path := filepath.Join(TaskDir, filename)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	entries, err := os.ReadDir(TaskDir)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", TaskDir, err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		withoutExt := strings.TrimSuffix(name, ".md")
		parts := strings.SplitN(withoutExt, "-", 2)
		idPart := parts[0]
		slugPart := ""
		if len(parts) == 2 {
			slugPart = parts[1]
		}
		if strings.HasPrefix(idPart, ref) ||
			strings.EqualFold(slugPart, ref) ||
			strings.Contains(slugPart, ref) {
			return filepath.Join(TaskDir, name), nil
		}
	}

	return "", fmt.Errorf("no task found matching %q", ref)
}
