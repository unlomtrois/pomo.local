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
2. `pomo doro` (`doro.go`) POSTs to the daemon via `internal/client`. If no
   daemon is reachable, `ensureDaemon` (`ensure.go`) **auto-spawns** a detached
   `pomo daemon` (`Setsid`, output redirected to `$XDG_STATE_HOME/pomo/daemon.log`)
   and waits for `/healthz`.
   The daemon also **serves the embedded Svelte UI** (`internal/webui`, `go:embed`)
   on `/` with SPA fallback, so one binary is CLI + API + timer + dashboard/calendar,
   same-origin. Build the UI into the binary with `make ui` (SvelteKit static build
   → synced into `internal/webui/dist`, which is gitignored except a placeholder)
   then `make build`.
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

- `init` — marks the current directory as a project ("git for time tracking"): creates a `.pomo/` with `config.json` (`{id, name}`) and a self-ignoring `.gitignore` (`*`) so it stays invisible to git. `internal/project` discovers the nearest `.pomo` by walking up from the cwd (like git finds `.git`); `doro`/`rest` tag the session with that project's id/name. The daemon upserts a `projects` row (keyed by the `.pomo` `ext_id`) and sets `sessions.project_id`. The daemon/DB remain the source of truth — `.pomo` only supplies project identity.
- `daemon` — the long-lived server. `--addr/-a` (default `client.DefaultAddr()`), `--verbose/-v`. The listen/dial address is `$POMO_ADDR` or `127.0.0.1:7420`; **both** the daemon and the CLI resolve it via `client.DefaultAddr()`, so setting `POMO_ADDR=:80` (with a one-time `sudo setcap`/`sysctl ip_unprivileged_port_start` grant) makes `http://pomo.local` reachable with no port. The client's `dialHost` normalizes a wildcard/empty host (`:80`) to `127.0.0.1` for connecting. Binding a privileged port without the grant fails with a hint (`privilegedPortHint`). `--mdns` advertises `<host>.local` (`--host`, default `pomo`) and binds all interfaces so it's reachable at e.g. `http://pomo.local:7420`. mDNS registers through the system avahi-daemon over D-Bus (`internal/mdns`, with the `NO_REVERSE` publish flag so the alias doesn't collide with the machine's own reverse PTR); falls back to an embedded pure-Go responder (`hashicorp/mdns`) only when avahi is absent.
- `doro [topic]` — start a fixed pomodoro via the daemon; auto-spawns it. Topic is a positional arg (optional, like `start`/`end`). `--email`, `--duration/-d`, `--message/-m`, `--hint`, `--verbose/-v`.
- `start [topic]` / `end [topic]` (alias `stop`) — open-ended **stopwatch**: `start` opens a session with no fixed stop and no timer (`sessions.open=1`); `end` records actual elapsed time, marks it done, and optionally sets/overrides the topic. Topic is a positional arg and optional on both sides (name at start, at end, or override). Enforced by the same single-active invariant as `doro`. `internal/store.EndActiveSession` computes elapsed = now − start; the daemon skips arming/reconciling timers for `open` sessions.
- `cancel` — cancel (discard) the active session — doro or open `start` — marking it `cancelled` (vs `end`/`stop` which records it `done`); disarms the timer via `POST /api/sessions/active/stop` (`client.StopActive`). Same effect as `active --remove`.
- `rest` — thin wrapper calling `executeStart("Rest", ...)` with a 5m default.
- `log` — git-log-style session history, newest first (`GET /api/sessions?project=<ext_id>&limit=N` → `store.ListSessionsByProject`). Scoped to the current `.pomo` project by default; `--all` lists across projects, `-n` limits. Prints `<marker> [hash] <dur> <when> <topic>` (marker: `▶` active, `✗` cancelled).
- `edit <hash>` — partial update of a past session by hash/prefix (`PATCH /api/sessions/by-hash/{prefix}` → `store.EditSession`). Flags `--topic/-t` (pass `""` to clear), `--start` (RFC3339 / `"2006-01-02 15:04"` / `"15:04"`), `--duration/-d`. Only passed flags change (detected via cobra `Flags().Changed`); keeps `stop = start + duration`; re-arms the timer if the active doro's timing changed.
- `rm <hash>` (alias `delete`) — delete a session by hash/prefix (`DELETE /api/sessions/by-hash/{prefix}` → `store.DeleteSession`); disarms the timer first if it's the active session.
- `show <hash>` — look up a session by its hash or a short prefix (git-style); the daemon resolves via `store.SessionByHashPrefix` (404 not-found / 409 ambiguous). Every session gets a stable random 40-hex `hash` at creation (`sessions.hash`, unique index; the git *object-name* analog — random, not content-derived, since sessions mutate). CLI output shows the 7-char short form `[abc1234]`.
- `status` (aliases `stats`, `health`) — shows DB file path/size (read locally), and — if the daemon is reachable — its version, session counts by status, project count, total tracked time, and the active session. Deliberately does **not** auto-spawn the daemon (`GET /api/stats` → `store.Stats`; the server carries the version passed to `server.New`).
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
