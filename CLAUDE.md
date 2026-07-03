# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Always use the Makefile targets — never raw `go build` (the build injects the version via ldflags):

- `make build` — builds the `pomo` binary with the git-derived version embedded into `main.version`
- `make install` — `go install` with the same ldflags
- `make test` — runs `go test ./...`
- `make lint` — runs `golangci-lint run ./...`
- `make clean` — removes the binary

Run a single test: `go test ./internal/utils/ -run TestShortDuration -v`

Linting uses `revive` plus `errcheck` (see `.golangci.yml`). Deferred `Close` calls are excluded from errcheck.

## Architecture

A Linux-targeted Pomodoro CLI. The key design is **self-re-invocation via the OS scheduler** rather than a long-lived process:

1. `pomo start` (`cmd/pomo/commands/start.go`) builds a `pomo.Session`, resolves its own binary path with `os.Executable()`, and constructs a `scheduler.Task` whose command is `pomo notify --summary ... --body ...`.
2. The task is handed to a `scheduler.Scheduler` (`internal/scheduler`) which schedules the *same binary* to run at the session's stop time, then the CLI exits immediately. No process stays alive during the timer.
3. When the timer fires, `pomo notify` (`notify.go`) sends the desktop notification, optionally emails, and cleans up the state files.

`scheduler.NewDefault` picks a backend at runtime: `SystemdScheduler` (transient `systemd-run --user` timer) if `systemd-run` is on PATH, else `AtScheduler` (pipes a `sleep && notify` script to `at`). Both implement the `Scheduler` interface; `Cancel` is currently unimplemented (`panic("todo")`).

### State and persistence (XDG dirs)

- `$XDG_STATE_HOME/pomo/active_session.json` — the running session; its existence enforces "only one active session at a time" (checked in `checkPomodoroSession`). `pomo active` reads it; `pomo notify` deletes it.
- `$XDG_STATE_HOME/pomo/active_task.json` — the scheduled `scheduler.Task`.
- `$XDG_DATA_HOME/pomo/sessions.csv` — append-only session history (topic, start, stop, duration).
- `$XDG_CONFIG_HOME/pomo/mail.json` — SMTP host/port/sender/receiver (`internal/config`).

### Commands (`cmd/pomo/commands`)

Cobra-based; each file registers itself onto `rootCmd` in an `init()`. `main.go` calls `SetVersion` then `Execute`.

- `start` — core flow above. `--email`, `--duration/-d`, `--topic/-t`, `--hint`, `--verbose/-v`.
- `rest` — thin wrapper calling `executeStart("Rest", ...)` with a 5m default.
- `notify` — the timer callback; not usually run by hand.
- `active` — inspects the active session; `--remove` clears a stale one.
- `auth` — stores SMTP/Toggl secrets in the OS keyring (`go-keyring`) and writes `mail.json`; for `--email` it live-tests the SMTP connection.

### Notifications

- Desktop: `internal/notifier` shells out to `notify-send`.
- Email: `internal/mail` loads `MailConfig`, pulls the app password from the keyring (`pomo-smtp` service), and uses `net/smtp`.
- Secrets never live in config files — only the keyring holds passwords/tokens.

### Incomplete / stubbed

- `internal/toggl` — Toggl Track API client exists (`api.go`) but the integration wiring (`toggl.go`) is `panic("todo")`. The `--toggl` flag is accepted but unused.
- `internal/storage/csv.go` — `InitCsv` exists but session appending is done inline in `start.go` (`appendSessionCsv`), not through this package.

## Conventions

- Module path is `pomo.local` (local module, not a real remote path).
- External processes (`systemd-run`, `at`, `notify-send`) are invoked via `os/exec`; availability is probed with `exec.LookPath`.
- Debug logging uses `log/slog`; `--verbose` raises the level to `Debug`.
