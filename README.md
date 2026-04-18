# grit

> Flat-file task tracker. Sharp and direct.

`grit` is a CLI task manager where every task is a Markdown file. No database.
No daemon. No sync service. Tasks live in `.grit/` alongside your code - commit
them, grep them, read them with any editor.

---

## Install

```bash
git clone https://github.com/Salvadego/grit
cd grit
go build -o grit .
mv grit /usr/local/bin/

# you can also do something like:
go build -o $GOPATH/bin/grit .
```

Or with `go install`:

```bash
go install github.com/Salvadego/grit@latest
```

---

## Quick start

```bash
grit init                                    # creates .grit/
grit add "Fix the auth bug" -t bug,urgent
grit add "Write release notes" -t docs
grit list
grit close auth
grit list -q "@open"
```

---

## How it works

### Every task is a file

Running `grit add "Fix the auth bug"` creates:

```
.grit/1776481293919-fix-the-auth-bug.md
```

```markdown
---
id: 1776481293919
status: open
tags: [bug, urgent]
---
# Fix the auth bug
```

The filename is `<millisecond-timestamp>-<slug>.md` - human-readable, sortable,
collision-resistant without a central counter.

### The state file

`.grit/.state.bin` is a binary index of all tasks. It stores title, status,
tags, slug, and mtime for every file. On `grit list`, only files whose `mtime`
changed since the last run are re-parsed. Everything else is served directly
from the binary cache - no Markdown parsing, no file opens.

**State file layout:**

```
[Magic: 4 bytes "GriT"] [Version: 1 byte] [Count: 4 bytes]
[StateEntry x Count]
```

Each `StateEntry` is a fixed-width 914-byte record:

```
ID         uint64
MTime      int64
StatusByte uint8
TagCount   uint8
Title      [256]byte
Slug       [128]byte
Tags       [8][64]byte
```

On magic or version mismatch the state file is discarded and rebuilt from
scratch automatically.

---

## Commands

### `grit init`

Initializes grit in the current directory by creating `.grit/`.

```bash
grit init
```

---

### `grit add <title>`

Creates a new task file.

```bash
grit add "Refactor the parser"
grit add "Deploy to staging" -t devops,urgent
```

| Flag | Description |
|------|-------------|
| `-t, --tags` | Comma-separated tags |

---

### `grit list`

Lists tasks. Served from the state cache - disk access only for files that
changed.

```bash
grit list                          # all tasks
grit list -q "@open"               # open tasks only
grit list -q ".bug"                # tasks tagged 'bug'
grit list -q ".bug AND @open"      # open bug tasks
grit list -q "title ~ 'auth'"      # regex on title
grit list -q "NOT @closed"         # everything not closed
grit list -q "id >= 1776481000000" # by ID range
grit list -q "tags?"               # tasks that have any tag
```

|     Flag      |                        Description                        |
|    ------     |                       -------------                       |
| `-q, --query` | Filter expression (see [Query Language](#query-language)) |

---

### `grit edit <ref>`

Opens a task in `$EDITOR` (falls back to `vi`).

```bash
grit edit 1776481293919       # by ID
grit edit fix-the-auth-bug    # by slug
grit edit auth                # by slug substring
```

---

### `grit close <ref>`

Marks a task as closed. Atomic write via tmp->rename.

```bash
grit close 1776481293919
grit close auth
```

---

### `grit reopen <ref>`

Reopens a closed task.

```bash
grit reopen auth
```

---

### `grit delete <ref>`

Permanently deletes a task file.

```bash
grit delete auth
grit rm 1776481293919
```

---

## Resolving tasks by `<ref>`

All commands that operate on a single task accept a `<ref>` argument.
Resolution order:

1. **State cache** - match by ID prefix or slug (no disk scan)
2. **Filesystem fallback** - `ReadDir` scan for cold start or cache miss

A `<ref>` matches if:
- It is a prefix of the numeric ID (`1776481` matches `1776481293919`)
- It exactly matches the slug (`fix-the-auth-bug`)
- It is a substring of the slug (`auth` matches `fix-the-auth-bug`)

---

## Query language

`grit list` uses [qlvm](https://github.com/Salvadego/qlvm) - an embeddable
stack-based query VM.

### Fields

|   Field    |   Type   |        Example        |
| ---------- | -------- |       ---------       |
|  `title`   |  string  |   `title ~ 'auth'`    |
|  `status`  |  string  |   `status = 'open'`   |
|   `tags`   |  string  | `tags contains 'bug'` |
|    `id`    |  number  | `id >= 1776481000000` |

### Operators

|   Operator   |              Meaning               |
| ------------ |             ---------              |
|     `=`      |     equals (case-insensitive)      |
|     `!=`     |             not equals             |
|  `>`  `>=`   |      greater than / or equal       |
|  `<`  `<=`   |        less than / or equal        |
|     `~`      |            regex match             |
|  `contains`  | substring match (case-insensitive) |

### Logic

```
AND   OR   NOT   ( )
```

### Prefix symbols

|  Symbol  |      Expands to       |
| -------- |      -----------      |
|  `.bug`  | `tags contains 'bug'` |
| `@open`  |   `status = 'open'`   |

### Suffix symbols

|  Symbol  |         Expands to         |
| -------- |        -----------         |
| `tags?`  | `tags != ''` (has any tag) |
| `title?` |       `title != ''`        |

### Examples

```bash
grit list -q "@open AND .bug"
grit list -q "title ~ 'refactor|cleanup'"
grit list -q "NOT (@closed OR .wontfix)"
grit list -q ".urgent AND id < 1776500000000"
```

---

## File layout

```
your-project/
└── .grit/
    ├── .state.bin                        <- binary cache index
    ├── 1776481293919-fix-the-auth-bug.md
    ├── 1776481439085-write-release-notes.md
    └── ...
```

Add `.grit/` to version control to share tasks with your team. Add
`.grit/.state.bin` to `.gitignore` - it's a local cache and is always
rebuildable.

```gitignore
.grit/.state.bin
.grit/.state.bin.tmp
```

---

## Configuration

`grit` respects the `EDITOR` environment variable for `grit edit`. If unset,
falls back to `vi`.

A config file at `$HOME/.config/grit.yaml` is loaded automatically by Viper if
present:

```yaml
editor: nvim
```
