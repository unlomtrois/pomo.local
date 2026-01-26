# feature
- [x] init pomodoro csv - with headers
- [x] move saving pomodoro csv into separate method

# refactoring
- [x] move main to cmd
- [x] move pomodoro struct to internal

# feature

## toggl integration
- [x] fn to add a new entry
- [x] move it from main to internal
- [x] integrate it to main `pomo` package

# feature
- [x] add `--version` flag

# feature
- make -d accept duration in seconds, minutes, hours, e.g. 3600s, 60m, 1h

# backlog
- mdns server, maybe daemon
- obsidian markdown integration (provide obsidian links where to export?)
- export command
- `current` command to get active entry
    - (toggle) fn to get current entry
- `track` command, `stop` command, with `--toggl` flag
- add `--markdown <pomodoro.md>` flag to export to markdown
- add playing sound on notify, and --mute flag to disable sound
- make `start` command to accept positional string arg, e.g. `pomo start "focus"`