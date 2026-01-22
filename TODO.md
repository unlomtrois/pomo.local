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

# backlog
- simple args parser to be able to set -d in seconds, hours, e.g. 600s, 1h
- mdns server, maybe daemon
- obsidian markdown integration (provide obsidian links where to export?)
- export command
- `current` command to get active entry
    - (toggle) fn to get current entry
- `track` command, `stop` command, with `--toggl` flag