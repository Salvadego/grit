---
id: 1776488687661
status: closed
tags: [feature, v0.2.0, v0.3.0, state]
relations: []
---
# remove limits from the state

This will be a multipart plan...
The fixed-width StateEntry struct is what enables the binary format....
`encoding/binary` requires fixed sizes, so I can't just remove the
limits without changing the storage strategy.

First I will start by increasing the capacity to a higher number such as:
```
const maxTags      = 32
const maxRelations = 32
const StateVersion = uint8(3)
```

If this ever becomes insufficient, the layout would need to change...
The idea is to split the state file into two "zones":

`.grit/.state.bin` - fixed-width index, one record per task, fast scan:
```
[Header]
[CoreEntry x N]: ID, MTime, StatusByte, Title[256], Slug[128] -> (~410 bytes each)
```

`.grit/.state.ext` - variable-length extension data, keyed by ID:
`[ID uint64][TagCount uint16][tags...][RelCount uint16][rels...]`

