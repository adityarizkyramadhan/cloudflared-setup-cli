# Cross-Platform Support (Windows + Linux Auto-Detect) — Design

**Date:** 2026-06-18
**Status:** Approved

## Problem

The CLI currently only works on Linux. The primary blocker is a hardcoded
install path (`/usr/local/bin`) in the auth screen, plus orchestration that
forces the user to manually pick between systemd and Windows Service and always
prints `sudo` instructions. The user runs the tool directly on a Windows server
and wants it to auto-detect the host OS (`runtime.GOOS`) and adapt.

## Goals

1. Make `install / download cloudflared` work on Windows (and stay working on Linux).
2. Auto-detect the host OS and pick the right install directory and service manager.
3. Auto-elevate to admin on Windows when an operation requires it.

## Non-Goals (YAGNI)

- SSH / remote VM management. The CLI runs on the machine it configures.
- MSI installer integration.
- Linux auto-elevation via `sudo` re-exec (show a clear message instead).

## Deployment Model

The CLI runs directly on the server it configures. "Auto-detect VM OS" means
detecting the local OS via `runtime.GOOS` — no remote probing.

## Architecture

### New package: `internal/platform/`

Centralizes every OS-specific decision. Uses build-tagged files so each OS
compiles cleanly:

- `platform_windows.go` (`//go:build windows`)
- `platform_unix.go` (`//go:build !windows`)
- `platform_test.go` (cross-platform assertions keyed off `runtime.GOOS`)

Exported API (same signatures in both build files):

| Function | Windows behavior | Unix behavior |
|----------|------------------|---------------|
| `InstallDir() (string, error)` | First writable of: dir of existing `cloudflared` on PATH → `%ProgramData%\cloudflared` → `%LOCALAPPDATA%\cloudflared\bin` → `%USERPROFILE%\.cloudflared\bin`. Created if missing. | First writable of: `/usr/local/bin` → `~/.local/bin` → `~/.cloudflared/bin`. Created if missing. |
| `ServiceManager() string` | `"windows"` | `"systemd"` on Linux, `"launchd"` on darwin, else `"none"` |
| `IsAdmin() bool` | Token elevation check via `golang.org/x/sys/windows` | `os.Geteuid() == 0` |
| `RelaunchElevated() error` | `ShellExecute` verb `runas` on `os.Executable()` with the same args (triggers UAC) | no-op returning an error that signals "run with sudo" |

`InstallDir()` "auto-detect to all possible locations": iterate candidates,
return the first that is writable (probe by creating the directory and writing
+ removing a temp marker file). This is the user's chosen behavior.

### Dependency

Promote `golang.org/x/sys` from indirect to direct (already present at
`v0.38.0` in `go.sum`). Used for `windows.GetCurrentProcessToken().IsElevated()`
and `windows.ShellExecute`.

## Changes to Existing Code

### 1. `ui/auth.go` — fix the primary blocker

`downloadCloudflared()` currently calls `cloudflared.Install("/usr/local/bin")`.
Replace with:

```
dir, err := platform.InstallDir()
// ... Install(dir)
```

Success message reports the real install path. On Windows, if the chosen dir is
not already on PATH, append a one-line reminder (no automatic PATH mutation in
this iteration — keep scope tight).

### 2. `ui/orchestration.go` — auto-detect + auto-elevate

- Option `[1]` becomes **"Install service native (auto-detect OS)"**.
  After the user enters the tunnel name:
  - `switch platform.ServiceManager()`:
    - `"windows"`: if `!platform.IsAdmin()` → call `platform.RelaunchElevated()`;
      on success return `tea.QuitMsg{}` (the elevated copy continues in a new
      console). On UAC cancel / error, show the error and stay (do not quit).
      If already admin → resolve the cloudflared path (`exec.LookPath`, fallback
      `filepath.Join(InstallDir(), "cloudflared.exe")`) and call
      `orchestration.InstallWindowsService(name, path)`.
    - `"systemd"`: call `orchestration.InstallSystemd(name)` directly. If it
      fails on permission, surface a clear "jalankan dengan sudo" message.
    - other: report unsupported.
- Options `[2]` Docker and `[3]` Kubernetes remain manual file generators.
- Remove the systemd success message that always printed `sudo cp ... systemctl`
  instructions.

### 3. `ui/maintenance.go` — fix POSIX path

`doBackup` builds `home + "/.cloudflared"`. Replace with
`cloudflared.ConfigDir()` (already cross-platform via `os.UserHomeDir` +
`filepath.Join`).

## Data Flow (Windows native-service install)

```
User picks [1] → enters tunnel name
  → ServiceManager() == "windows"
    → IsAdmin()?
        no  → RelaunchElevated() → UAC prompt
                 accept → spawn elevated copy, current TUI quits
                 cancel → show error, stay in TUI
        yes → InstallWindowsService(name, cloudflaredPath)
                 → sc create / sc description / sc start
                 → success / error message in TUI
```

## Error Handling

- `InstallDir()` returns an error only if *no* candidate is writable; the UI
  surfaces it.
- `RelaunchElevated()` failure (incl. UAC cancellation) is non-fatal: the
  current TUI keeps running and shows the error.
- Service install failures bubble up the underlying `sc` / `systemctl` stderr.

## Testing

- `internal/platform/platform_test.go`:
  - `ServiceManager()` matches the expected value for `runtime.GOOS`.
  - `InstallDir()` returns a non-empty path to a directory that exists and is
    writable.
- Existing tests across `internal/...` stay green.
- CI build matrix in `.github/workflows/build.yml` already covers Windows and
  Linux, exercising both build-tag paths at compile time.
