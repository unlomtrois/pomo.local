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
- [x] make -d accept duration in seconds, minutes, hours, e.g. 3600s, 60m, 1h

# feature
## playing sounds
- [x] add default hint for notify-send
- [x] add `--mute` flag to disable sounds, overrides default hint
- [x] add `--notify-sound` that overrides default hint with `string:sound-file:<filepath>`

# feature
- [x] make `start` & `rest` commands to accept positional arg, e.g. `pomo start "focus"` and `pomo rest "break"`

# fix
- [x] fix `pomo --help` printing only `--version` help 

# feat
- [x] add `sleep` into pipe to support second precision

# feat
- [x] save configs into ~/.local

---

# backlog
- add mdns server, maybe daemon
- add obsidian markdown integration (provide obsidian links where to export?)
- add `export` command
- add `current` command to get the active entry
- add `--markdown <pomodoro.md>` flag to export to markdown
- pomo `notify` command that talks with libnotify libnotify directly, and executable schedules itself `pomo notify ... | at ...`
- consider using [systemd/Timers](https://wiki.archlinux.org/title/Systemd/Timers) for sub-minute precision instead of `at` (which supports only minute-precision jobs)
- add `pomo toggl auth` command to fill toggl config to not ask for toggl-related flags
- I don't like that adding options after positional string does no work. e.g. `pomo start "somethind" -d 15m` -> -d 15 does not work and duration is defaulted to 25minutes
- add `-v` or `--verbose` flag to log what is exactly happening
- deslopify main cli
- write pomo logs in `~/.local/share/pomo`
- what if skipping `-t`, like `pomo start` would use the last used one instead of the default? or it could be `pomo continue`
