---
id: 1776488795926
status: closed
tags: [feature, v0.2.0, schema]
relations: [1776488920769:related]
---
# add creation/modified date fields to schema

The idea to add those fields to that filtering for recently changed will be used...
It would also be good if I could order the results by those fields...

```bash
grit list '@open' --sort modified      # newest modified first
grit list '@open' --sort created:asc   # oldest first
grit list 'created > 2026-03-01'       # created ones after the first of march
```
