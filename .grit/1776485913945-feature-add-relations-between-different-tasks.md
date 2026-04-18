---
id: 1776485913945
status: closed
tags: [feature, v0.2.0, tasks, state, query_engine]
relations: [1776488607037:blocked-by]
---
# feature: add relations between different tasks

The idea is to add relations / references between different tasks

## Example: 

```bash
grit add 'feature: update qlvm schema to include relation' -r 1776485913945 -t feature,v0.2.0,tasks,query_engine
```

## This includes some questions>
    1. Should the relations be updated? and have a new field such as mentioned?
    2. Should there be different types of relations? (parent/child, depency, mentioned, related, blocked by/blocks)


## Design:

|  Type   |  Inverse   |
|  ----   |  -------   |
| parent  |   child    |
| blocks  | blocked-by |
| related |  related   |

I will keep relations capped at 8 for now, with the hope to later remove such limits...

### qlvm schema:
Add relations as a string field (joined IDs) so you can query 
```bash
grit list -q "blocked-by?"
# or
grit list -q "relations contains '1776485'"
```
