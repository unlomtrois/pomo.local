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

# feat
- [x] if we use unknown command, cli tell about it

# feat
- pomo `notify` command and executable schedules itself `pomo notify ... | at ...` instead of calling notify-send
- [x] add `notify` command
- [x] replace `notify-send` with `pomo notify` 

# feat
- ~~consider using [systemd/Timers](https://wiki.archlinux.org/title/Systemd/Timers) for sub-minute precision instead of `at` (which supports only minute-precision jobs)~~
- use `systemd-run` instead of `at` which support one-time `systemd/Timers`
- [x] replace `at` with systemd-run

# feat
- add fallback to `at` for systems without systemd

# refactor
- [x] deslopify main cli

# refactor
- [x] make pomo "rest" command just an alias to more robust command (essentially 'start')

# refactor
- separate "summary" and "title". Title of pomo is something you're doing, and summary should just be "pomodoro session ended"

# feat
- [x] add `pomo auth` command that can take --mail / --toggl flags
- [x] uses https://pkg.go.dev/golang.org/x/term#Terminal.ReadPassword to read password
- [x] uses keyring/dbus, first dependency :(

# feat
- [x] add mail service
- [x] `pomo notify --mail` will send an email that the pomo session is ended
- [x] adding `--mail` in `pomo start` will call `pomo notify` with `--mail` flag

# refactor
- use [xdg](https://github.com/adrg/xdg) instead of declaring variables in utils

# feat
- if --email flag is provided, scheduler should also inform that it will notify via email as well

# feat
- write pomo sessions in `~/.local/share/pomo/sessions.csv`

# feat
- add `active` command to get the active pomodoro session

# feat
- add `-v --verbose` flag to log what is exactly happening, e.g. for these systemd-run timer stuff

# feat
- add shell.nix and default.nix, to make it work on nixos

# feat
- pomo notify erases the active session & active task in .local/state

---

# plan
- add `list` command to make systemctl calls to see all pending pomodoro timers

---

# backlog
- add mdns server, maybe daemon
- add obsidian markdown integration (provide obsidian links where to export?)
- add `export` command
- add `--markdown <pomodoro.md>` flag to export to markdown
- add `pomo toggl auth` command to fill toggl config to not ask for toggl-related flags
- what if skipping `-t`, like `pomo start` would use the last used one instead of the default? or it could be `pomo continue`
- `--strict` flag that allows only a single pomodoro session to be tracked
- replace keyring with my own package around godbus/dbus
- add `stop` command
