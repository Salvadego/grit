---
id: 1776522097181
status: closed
tags: [feature, v1.1.0, output, scripting]
relations: []
---
# render output based on context (piped vs terminal)

When stdout is a terminal, render the existing tablewriter table with colours.
When stdout is piped, render TSV - one task per line, header on line 1, no
ANSI codes, no box drawing.

## Commands affected
- `grit list`   - TSV instead of tablewriter
- `grit add`    - print only the new task ID
- `grit close`  - print only the task ID
- `grit reopen` - print only the task ID
- `grit delete` - print only the task ID
- `grit list -x --confirm` - one ID per line, no summary

## Commands NOT affected
- `grit edit`   - interactive, meaningless to pipe
- `grit init`   - one-time setup
- `grit relate` - no scripting value
- `grit unrelate`
