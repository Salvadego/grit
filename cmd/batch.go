package cmd

import (
	"fmt"
	"os"
	"strings"
)

type batchAction func(path string, arg string) error

var batchActions = map[string]batchAction{
	"close":  func(path string, _ string) error { return setStatusByPath(path, "open", "closed") },
	"reopen": func(path string, _ string) error { return setStatusByPath(path, "closed", "open") },
	"delete": func(path string, _ string) error { return deletePath(path) },
	"tag":    addTagToPath,
	"untag":  removeTagFromPath,
}

func execBatch(tasks []Task, rawAction string, confirm bool) error {
	verb, arg, _ := strings.Cut(rawAction, ":")
	verb = strings.ToLower(strings.TrimSpace(verb))
	arg = strings.TrimSpace(arg)

	fn, ok := batchActions[verb]
	if !ok {
		return fmt.Errorf("unknown action %q - available: close, reopen, delete, tag:<name>, untag:<name>", rawAction)
	}
	if (verb == "tag" || verb == "untag") && arg == "" {
		return fmt.Errorf("%s requires a tag name, e.g. -x %s:v0.3.0", verb, verb)
	}
	if len(tasks) == 0 {
		fmt.Println("grit: no matching tasks.")
		return nil
	}

	if !confirm {
		fmt.Printf("grit: %s will affect %d task(s):\n\n", rawAction, len(tasks))
		for _, t := range tasks {
			fmt.Printf("  %-16d  %s\n", t.ID, t.Title)
		}
		fmt.Printf("\ngrit: dry run - pass --confirm to apply\n")
		return nil
	}

	cache := LoadState()
	cacheModified := false
	var errs []string
	okCount := 0

	for _, t := range tasks {
		if err := fn(t.Filename, arg); err != nil {
			errs = append(errs, fmt.Sprintf("  %d: %v", t.ID, err))
			continue
		}
		switch verb {
		case "delete":
			delete(cache, t.ID)
			cacheModified = true
		case "close", "reopen":
			if entry, exists := cache[t.ID]; exists {
				entry.StatusByte = statusByte(map[string]string{
					"close":  "closed",
					"reopen": "open",
				}[verb])
				cache[t.ID] = entry
				cacheModified = true
			}
		case "tag", "untag":
			if entry, exists := cache[t.ID]; exists {
				entry.MTime = 0
				cache[t.ID] = entry
				cacheModified = true
			}
		}
		okCount++
		if isPiped() {
			fmt.Printf("%d\n", t.ID)
		} else {
			fmt.Printf("  %-16d  %s\n", t.ID, t.Title)
		}
	}

	if cacheModified {
		_ = SaveState(cache)
	}

	if !isPiped() {
		fmt.Printf("\ngrit: %d/%d succeeded\n", okCount, len(tasks))
	}
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "grit: errors:\n%s\n", strings.Join(errs, "\n"))
	}

	return nil
}

func setStatusByPath(path, from, to string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	body := string(raw)
	if !strings.Contains(body, "status: "+from) {
		return fmt.Errorf("already %s", to)
	}
	body = strings.Replace(body, "status: "+from, "status: "+to, 1)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(body), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func deletePath(path string) error {
	return os.Remove(path)
}

func addTagToPath(path, tag string) error {
	t, err := ParseTask(path)
	if err != nil {
		return err
	}
	for _, existing := range t.Tags {
		if strings.EqualFold(existing, tag) {
			return nil
		}
	}
	t.Tags = append(t.Tags, tag)
	return rewriteFrontmatter(path, t)
}

func removeTagFromPath(path, tag string) error {
	t, err := ParseTask(path)
	if err != nil {
		return err
	}
	filtered := t.Tags[:0]
	for _, existing := range t.Tags {
		if !strings.EqualFold(existing, tag) {
			filtered = append(filtered, existing)
		}
	}
	if len(filtered) == len(t.Tags) {
		return nil
	}
	t.Tags = filtered
	return rewriteFrontmatter(path, t)
}
