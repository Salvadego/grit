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

# or:
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
grit list "@open"
```

---

## How it works

### Every task is a file

Running `grit add "Fix the auth bug" -t bug,urgent` creates:

```
.grit/1776481293919-fix-the-auth-bug.md
```

```markdown
---
id: 1776481293919
status: open
tags: [bug, urgent]
relations: []
---
# Fix the auth bug
```

The filename is `<millisecond-timestamp>-<slug>.md` - human-readable, sortable,
collision-resistant without a central counter.

### The state file

`.grit/.state.bin` is a binary index of all tasks. It stores title, status,
tags, relations, slug, and mtime for every file. On `grit list`, only files
whose `mtime` changed since the last run are re-parsed. Everything else is
served directly from the binary cache - no Markdown parsing, no file opens.

**State file layout:**

```
[Magic: 4 bytes "GRIT"] [Version: 1 byte] [Padding: 1 byte] [Count: 8 bytes]
[StateEntry x Count]
```

Each `StateEntry` is a fixed-width record:

```
ID             uint64
MTime          int64
StatusByte     uint8
TagCount       uint8
RelationCount  uint8
_              [5]uint8  (padding)
Title          [256]byte
Slug           [128]byte
Tags           [32][64]byte
Relations      [32]RelationEntry
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

Creates a new task file. Prints the new task ID to stdout - scriptable.

```bash
grit add "Refactor the parser"
grit add "Deploy to staging" -t devops,urgent
grit add "Fix login" -t bug -r 1776481293919:blocks

# capture the ID in a script
id=$(grit add "hotfix: nil deref" -t bug)
grit relate "$id" 1776481293919 blocks
```

|       Flag        |                              Description                               |
|      ------       |                             -------------                              |
|   `-t, --tags`    |                          Comma-separated tags                          |
| `-r, --relations` | Related task IDs, e.g. `-r 1776481293919` or `-r 1776481293919:blocks` |

---

### `grit list [query]`

Lists tasks. Served from the state cache - disk access only for files that
changed. Defaults to showing open tasks.

```bash
grit list                              # open tasks (default)
grit list "@open"                      # open tasks
grit list ".bug"                       # tasks tagged 'bug'
grit list ".bug AND @open"             # open bug tasks
grit list "title ~ 'auth'"             # regex on title
grit list "NOT @closed"                # everything not closed
grit list "id >= 1776481000000"        # by ID range
grit list "tags?"                      # tasks that have any tag
grit list "created > 2026-03-01"       # created after date
grit list "modified > 2026-04-01"      # modified after date
grit list "@open" --sort modified      # newest modified first
grit list "@open" --sort created:asc   # oldest first
```

**Batch operations:**

```bash
grit list ".bug AND @open" -x close            # dry run
grit list ".bug AND @open" -x close --confirm  # apply
grit list "@open" -x tag:v0.2.0 --confirm      # add tag to all matched
grit list ".v0.1.0" -x untag:v0.1.0 --confirm  # remove tag from all matched
grit list "@closed" -x delete --confirm         # delete all closed tasks
```

**Terminal output** adapts to the available width:

```
<  100 cols  ->  id, status, title
>= 100 cols  ->  + tags
>= 130 cols  ->  + relations
>= 150 cols  ->  + created
>= 170 cols  ->  + modified
```

**Piped output** is TSV - one task per line, header on line 1, no ANSI codes.
Safe for `awk`, `cut`, `jq`, `mlr`. Relations are pipe-separated within their
field: `blocks:1776481293919|blocked-by:1776481439085`

```
id\tstatus\ttitle\ttags\trelations\tcreated\tmodified
```

See [Scripting](#scripting) for examples.

|     Flag      |                                Description                                 |
|    ------     |                               -------------                                |
| `-s, --sort`  | Sort field and direction: `created`, `modified` (append `:asc` or `:desc`) |
| `-x, --exec`  |     Batch action: `close`, `reopen`, `delete`, `tag:<n>`, `untag:<n>`      |
|  `--confirm`  |          Confirm and apply the batch action (default is dry run)           |
| `-q, --query` |         Filter expression - deprecated, use positional arg instead         |

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

Permanently deletes a task file and removes it from the state cache.

```bash
grit delete auth
grit rm 1776481293919
```

---

### `grit relate <ref> <target-id> [type]`

Adds a typed relation between two tasks. The inverse relation is written to
the target file automatically.

```bash
grit relate auth 1776481293919             # related (default)
grit relate auth 1776481293919 blocks      # this task blocks target
grit relate auth 1776481293919 blocked-by
grit relate auth 1776481293919 parent
grit relate auth 1776481293919 child
```

|  Relation type  |   Inverse    |
| --------------- |  ---------   |
|    `parent`     |   `child`    |
|     `child`     |   `parent`   |
|    `blocks`     | `blocked-by` |
|  `blocked-by`   |   `blocks`   |
|    `related`    |  `related`   |

---

### `grit unrelate <ref> <target-id>`

Removes a relation between two tasks. Also removes the inverse.

```bash
grit unrelate auth 1776481293919
```

---

### `grit completion [bash|zsh|fish]`

Generates shell completion scripts. Tab-completes task slugs, IDs, query
fields, prefix symbols, tags, and batch actions.

```bash
# bash - add to ~/.bashrc
source <(grit completion bash)

# zsh - add to ~/.zshrc
source <(grit completion zsh)

# fish
grit completion fish | source
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

|    Field     |  Type  |                Example                |
|   -------    | ------ |               ---------               |
|   `title`    | string |           `title ~ 'auth'`            |
|   `status`   | string |           `status = 'open'`           |
|    `tags`    | string |         `tags contains 'bug'`         |
|     `id`     | number |         `id >= 1776481000000`         |
| `relations`  | string | `relations contains '1776481293919'`  |
|  `blocked`   |  bool  |               `blocked`               |
|   `blocks`   | string |               `blocks?`               |
| `blocked-by` | string | `blocked-by contains '1776481293919'` |
|   `parent`   | string |               `parent?`               |
|   `child`    | string |               `child?`                |
|  `related`   | string |              `related?`               |
|  `created`   |  date  |        `created > 2026-03-01`         |
|  `modified`  |  date  |        `modified > 2026-04-01`        |

### Operators

|  Operator  |              Meaning               |
| ---------- |             ---------              |
|    `=`     |     equals (case-insensitive)      |
|    `!=`    |             not equals             |
|  `>` `>=`  |      greater than / or equal       |
|  `<` `<=`  |        less than / or equal        |
|    `~`     |            regex match             |
| `contains` | substring match (case-insensitive) |

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

|  Symbol   |            Expands to             |
| --------  |            -----------            |
|  `tags?`  |    `tags != ''` (has any tag)     |
| `title?`  |           `title != ''`           |
| `blocks?` | `blocks != ''` (blocks something) |
| `parent?` |   `parent != ''` (has a parent)   |

### Examples

```bash
grit list "@open AND .bug"
grit list "title ~ 'refactor|cleanup'"
grit list "NOT (@closed OR .wontfix)"
grit list ".urgent AND id < 1776500000000"
grit list "blocked"
grit list "@open AND blocks?"
grit list "created > 2026-03-01 AND @open"
grit list "@open" --sort modified:desc
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

---

## Versioning

```
dev -> git tag v0.x.y -> go install @latest
                               v
                       goreleaser builds binaries
                       and attaches to GitHub release
```

---

## Scripting

```bash
# capture new task ID
id=$(grit add "hotfix: nil deref in parser" -t bug)
grit relate "$id" 1776481293919 blocks

# close all v1.0.0 tasks, capture IDs
grit list ".v1.0.0 AND @open" -x close --confirm | while read -r id; do
    echo "closed: $id"
done

# count open tasks per tag
grit list "@open" | awk -F'\t' '
NR == 1 { next }
{ n = split($4, tags, ","); for (i=1;i<=n;i++) count[tags[i]]++ }
END { for (tag in count) printf "%3d  %s\n", count[tag], tag }
' | sort -rn

# just titles of open bug tasks
grit list ".bug AND @open" | cut -f3 | tail -n +2

# open every urgent task in editor
grit list "@open AND .urgent" | cut -f1 | tail -n +2 | xargs -I{} grit edit {}

# export open tasks to JSON
grit list "@open" | awk -F'\t' '
NR==1 { for(i=1;i<=NF;i++) hdr[i]=$i; next }
{ printf "{"; for(i=1;i<=NF;i++) printf "\"%s\":\"%s\"%s",hdr[i],$i,(i<NF?",":""); print "}" }
' | jq -s .

# fzf task picker
grit list "@open" | \
    awk -F'\t' 'NR>1 { printf "%s  %-40s  %s\n",$1,$3,$4 }' | \
    fzf --prompt="edit> " | awk '{print $1}' | xargs -r grit edit
```

---

## Changelog

### v1.1.0

- **Adaptive terminal output** - `grit list` detects terminal width and shows
  fewer columns on narrow terminals, more on wide ones.
- **Piped TSV output** - when stdout is piped, all commands emit
  machine-readable output: `grit list` writes TSV with a header row, mutating
  commands (`add`, `close`, `reopen`, `delete`) print only the affected task
  ID. Batch operations print one ID per line.

### v1.0.0

- Stable CLI contract declared. No functional changes from v0.2.0.

### v0.2.0

- **Relations** - typed task-to-task references (`blocks`, `blocked-by`,
  `parent`, `child`, `related`). Inverse relations are written to both files
  automatically. New commands: `grit relate`, `grit unrelate`. New query
  fields: `relations`, `blocked`, `blocks`, `blocked-by`, `parent`, `child`,
  `related`.
- **Query as positional arg** - `grit list '@open'` instead of
  `grit list -q '@open'`. The `-q` flag is kept as a hidden alias for
  backwards compatibility.
- **Shell completion** - tab-completes task slugs and IDs for `edit`, `close`,
  `reopen`, `delete`, `relate`, `unrelate`. Query completion for fields, prefix
  symbols (`@`, `.`), and live tag suggestions from the state cache.
- **Date fields** - `created` and `modified` fields in the query engine.
  Filter by creation or modification date. Sort results with `--sort modified`
  or `--sort created:asc`.
- **Batch operations** - `grit list <query> -x <action>`. Actions: `close`,
  `reopen`, `delete`, `tag:<n>`, `untag:<n>`. Dry run by default;
  `--confirm` to apply. State cache is updated atomically after each batch.
- **State limits raised** - `maxTags` and `maxRelations` increased from 8 to
  32. `StateVersion` bumped to 3; old state files are rebuilt automatically.

### v0.1.0

- Initial release.
- `grit init`, `grit add`, `grit list`, `grit edit`, `grit close`,
  `grit reopen`, `grit delete`.
- Binary state file with magic + version header for fast cache invalidation.
- qlvm query engine integration.
- Millisecond timestamp IDs with slug filenames.
- Atomic writes via tmp->rename throughout.

---

## License

MIT
