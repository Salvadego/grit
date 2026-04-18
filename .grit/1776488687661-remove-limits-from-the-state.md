---
id: 1776488687661
status: open
tags: [feature, v0.2.0, state]
relations: []
---
# remove limits from the state

This task is fairly simple...
All I need to do is to remove the limits from the state:
const maxTags = 8
const maxRelations = 8
const entrySize = 8 + 8 + 1 + 1 + 256 + 128 + 8*64

...

