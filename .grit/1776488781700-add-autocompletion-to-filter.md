---
id: 1776488781700
status: open
tags: [feature, v0.2.0, autocompletion]
relations: [1776488713202:related]
---
# add autocompletion to filter

This is a little bit more complicated...
This is kind of two changes:
1. Add the completions for the fields and some values, such as the relations to
   the query flag...
2. Migrate the flag to the first argument, the query flag is the only flag
   currently being used by the list command, which means -q is quite
   unnecessary (this may change since I plan to add batch
   operations later... such as closing multiple tasks at the same
   time)
