# Plain CLI Revamp + Bug Fixes — Design

**Date:** 2026-06-18
**Status:** Approved

## Problem

The Bubbletea TUI is considered too fiddly. Replace it with a plain
`fmt.Println`-based interactive menu CLI. While rewriting the UI layer, wire up
the features that are currently stubbed or half-implemented.

## Goals

1. Remove the Bubbletea/lipgloss TUI; drive the app with a plain interactive
   numbered-menu loop over stdin/stdout.
2. Fix the half-wired features so every menu item actually works.

## Non-Goals (YAGNI)

- Colors / styling / TUI widgets.
- Subcommand or flag-based CLI.
- Editing existing ingress rules.
- macOS `.tgz` extraction (guard with a clear message instead).

## Architecture

Rewrite the `ui` package (name kept so `cmd/root.go` barely changes) as a set
of plain functions. Remove `ui/model.go`, `ui/styles.go`, `ui/mainmenu.go` and
the Bubbletea model types in every screen file.

- **`ui/console.go`** — a `Console` struct wrapping `io.Reader`/`io.Writer` with
  a buffered reader, plus testable helpers:
  - `prompt(label) string`, `promptDefault(label, def) string`
  - `confirm(label) bool` (y/N; default no)
  - `readChoice() string` (reads a menu selection line)
  - print helpers: `info`, `ok` (✓), `fail` (✗), `warn` (⚠) via `fmt.Fprintln`
  - `Run(in io.Reader, out io.Writer) error` — entry point; `main` passes
    `os.Stdin`/`os.Stdout`. Returns on EOF (Ctrl+D / piped input) cleanly.
- **`ui/app.go`** — `Console.mainMenu()`: print the main menu, read a choice,
  dispatch to an area sub-loop, repeat until the user exits.
- **`ui/auth.go`, `credentials.go`, `monitoring.go`, `orchestration.go`,
  `maintenance.go`** — each exposes a method on `*Console` that runs that area's
  own submenu loop (print submenu → read choice → run action → print ✓/✗ →
  loop until "0" back).

`cmd/root.go` calls `ui.Run(os.Stdin, os.Stdout)` and maps errors to exit codes.

## Menus and Actions

**Auth & Setup**
1. Check install (`IsInstalled` + `GetVersion`)
2. Install/download (`platform.InstallDir` → `Install` → PATH reminder)
3. Login (`Login`)
4. Verify connection (`VerifyConnection`)

**Credentials**
1. List tunnels (`ListTunnels`)
2. Create tunnel (`CreateTunnel` + `SetTunnel` active)
3. Delete tunnel (`DeleteTunnelWithCleanup`)
4. Configure ingress (hostname + service → `AddIngressRule`, then auto
   `RouteDNS` using `ActiveTunnel`, prompting for tunnel if unset; RouteDNS
   failure is a non-fatal warning)
5. Route DNS (manual: hostname + tunnel → `RouteDNS`)
6. Export config (`store.BackupTo(path)`, default `./cloudflared-backup`)
7. Import config (`store.RestoreFrom(path)`) — newly wired

**Monitoring**
1. Live logs — stream `LogStreamer` lines to stdout; trap SIGINT so Ctrl+C stops
   streaming and returns to the menu instead of killing the app
2. Status (`GetStatus` of the active tunnel from config)
3. Metrics via API — prompt CF API token + account ID, `ValidateToken`, then
   list tunnels with status (the meaningful data the API client exposes)
4. Health check — prompt endpoint (default `http://localhost:8080`) → `CheckHealth`

**Orchestration**
1. Install native service — `ServiceManager` auto-detect; on Windows elevate via
   `RelaunchElevated` when not admin, else `InstallWindowsService`; on Linux
   `InstallSystemd`
2. Docker Compose (name + token → `DockerCompose`, write file)
3. Kubernetes manifest (name → `KubernetesManifest`, write file)

**Maintenance**
1. Update cloudflared — `LatestVersion` vs `CurrentVersion`; if newer, confirm
   then `UpdateBinary`
2. Cleanup orphaned credentials — list tunnels in use (`ListTunnels`), scan the
   config dir via `OrphanedCredentials`, show the orphans, confirm, `DeleteFiles`
3. Backup config (`CopyDir`/`store.BackupTo`)
4. Reset / uninstall (confirm "yes" → `UninstallCloudflared` + `RemoveConfigDir`)

## Live Logs Without a TUI

Start the `LogStreamer`, then loop `NextLine()` printing each line. Install a
`signal.Notify` handler for `os.Interrupt` for the duration of streaming: a
goroutine pushes lines to a channel; `select` on lines vs. the interrupt. On
interrupt or EOF, `streamer.Stop()`, reset signal handling, and return to the
menu.

## Bug Fixes Included

- Metrics, health-check endpoint input, cleanup, and import are wired (above).
- macOS install returns a clear "auto-install unsupported (.tgz)" error rather
  than writing a tarball to the binary path (mirrors the update guard).
- Remove the dead `credWaitingExportPath` constant (whole TUI state machine is
  removed anyway).

## Dependencies

`go mod tidy` drops `bubbletea`, `bubbles`, `lipgloss`, and their transitive
indirects. Remaining direct deps: `gopkg.in/yaml.v3`, `golang.org/x/sys`.

## Error Handling

Each action prints `✓`/`✗`/`⚠` and returns to its submenu. Invalid menu input
reprints the menu. Backend errors are surfaced verbatim. EOF on stdin exits the
program cleanly with code 0.

## Testing

- `ui/console_test.go` — drive `Console` with a `strings.Reader`:
  `confirm` (y/yes/blank/no), `promptDefault` (blank → default), `readChoice`.
- Existing `internal/...` tests stay green.
- `go build`/`go vet`/`go test ./...` clean on Windows; cross-compile
  linux/amd64, linux/arm64, darwin.
- The interactive loops and exec/stream actions are verified by build + manual
  run (consistent with existing untested exec/UI code).
