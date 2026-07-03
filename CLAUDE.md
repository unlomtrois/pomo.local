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

A Linux-targeted Pomodoro CLI evolving into a localhosted Toggl alternative. The
current design is a **long-lived daemon that owns all state**; the CLI (and later
the web UI and MCP server) are thin clients that talk to it over HTTP.

1. `pomo daemon` (`cmd/pomo/commands/daemon.go`) runs the server: it opens the
   SQLite store, serves the HTTP API (`internal/server`), and holds an
   in-process timer engine (`time.AfterFunc` per active session) that fires the
   completion notification/email itself. Intended to run under a systemd `--user`
   unit. Graceful shutdown on SIGINT/SIGTERM.
2. `pomo start` (`start.go`) POSTs to the daemon via `internal/client`. If no
   daemon is reachable, `ensureDaemon` (`ensure.go`) **auto-spawns** a detached
   `pomo daemon` (`Setsid`, output redirected to `$XDG_STATE_HOME/pomo/daemon.log`)
   and waits for `/healthz`.
3. When a session's timer fires, the daemon marks it `done` and calls the
   notifier (and `mail.SendMail` if the session requested email). On startup
   `Server.Reconcile` re-arms a still-running session's timer, or completes one
   that elapsed while the daemon was down.

The single-active-session invariant is enforced atomically by a **partial unique
index** (`sessions.status WHERE status='active'`) rather than a state-file check.

### State and persistence

- `$XDG_DATA_HOME/pomo/pomo.db` — SQLite, the daemon is the **sole writer**
  (`internal/store`). WAL mode; `sessions` and `projects` tables. Session rows
  carry the completion payload (`message`, `hint`, `email`) so the daemon's timer
  can reproduce the notification, plus `status` and `source` (cli/web/mcp).
- `$XDG_STATE_HOME/pomo/daemon.log` — stdout/stderr of an auto-spawned daemon.
- `$XDG_CONFIG_HOME/pomo/mail.json` — SMTP host/port/sender/receiver (`internal/config`).

### Commands (`cmd/pomo/commands`)

Cobra-based; each file registers itself onto `rootCmd` in an `init()`. `main.go` calls `SetVersion` then `Execute`.

- `init` — marks the current directory as a project ("git for time tracking"): creates a `.pomo/` with `config.json` (`{id, name}`) and a self-ignoring `.gitignore` (`*`) so it stays invisible to git. `internal/project` discovers the nearest `.pomo` by walking up from the cwd (like git finds `.git`); `start`/`rest` tag the session with that project's id/name. The daemon upserts a `projects` row (keyed by the `.pomo` `ext_id`) and sets `sessions.project_id`. The daemon/DB remain the source of truth — `.pomo` only supplies project identity.
- `daemon` — the long-lived server. `--addr/-a` (default `client.DefaultAddr`, `127.0.0.1:7420`), `--verbose/-v`. `--mdns` advertises `<host>.local` (`--host`, default `pomo`) and binds all interfaces so it's reachable at e.g. `http://pomo.local:7420`. mDNS registers through the system avahi-daemon over D-Bus (`internal/mdns`, with the `NO_REVERSE` publish flag so the alias doesn't collide with the machine's own reverse PTR); falls back to an embedded pure-Go responder (`hashicorp/mdns`) only when avahi is absent.
- `start` — daemon client; auto-spawns the daemon. `--email`, `--duration/-d`, `--topic/-t`, `--message/-m`, `--hint`, `--verbose/-v`.
- `rest` — thin wrapper calling `executeStart("Rest", ...)` with a 5m default.
- `active` — queries the daemon for the active session; `--remove` cancels it.
- `notify` — manual immediate-notification utility (no longer part of the timer flow).
- `auth` — stores SMTP/Toggl secrets in the OS keyring (`go-keyring`) and writes `mail.json`; for `--email` it live-tests the SMTP connection.

### Legacy / not on the main path

- `internal/scheduler` (`systemd-run`/`at` backends) — the pre-daemon
  self-re-invocation mechanism. Kept as an intended crash-survival fallback (fire
  a pending notification via the OS if the daemon dies); not currently wired in.
- `internal/storage/csv.go` and `internal/pomo` CSV serialization — superseded by
  the SQLite store; CSV is planned to return as an *export* format.

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
