# Cloudflared Setup CLI — Design Spec

**Date:** 2026-06-07  
**Author:** Aditya Rizky Ramadhan  
**Repository:** https://github.com/adityarizkyramadhan/cloudflared-setup-cli  
**Status:** Approved

---

## Overview

A terminal-based interactive CLI tool written in Go that works like an ATM — the user navigates entirely by pressing number keys, no need to memorize commands. The tool wraps the `cloudflared` binary and Cloudflare API to provide full tunnel lifecycle management: authentication, credential management, observability, multi-platform orchestration, and maintenance.

---

## Goals

- Provide a fully menu-driven TUI for managing Cloudflare Tunnels
- Work out-of-the-box on Linux, Windows, and macOS (amd64 + arm64)
- Installable as a single binary via `go install` or GitHub Releases download
- No prior knowledge of `cloudflared` CLI required to operate

---

## Non-Goals

- Does not re-implement cloudflared internals — wraps the binary
- Does not manage Cloudflare Workers, Pages, or other Cloudflare products
- Does not support multi-account switching in v1

---

## Tech Stack

| Concern | Choice |
|---|---|
| Language | Go |
| TUI Framework | `charmbracelet/bubbletea` |
| Styling | `charmbracelet/lipgloss` |
| UI Components | `charmbracelet/bubbles` (list, spinner, viewport) |
| Cloudflare API | `cloudflare/cloudflare-go` |
| Build & Release | GoReleaser + GitHub Actions |
| Credential Storage | `~/.cloudflared/` (local files) |
| Target Platforms | Linux, Windows, macOS — amd64 & arm64 |

---

## Project Structure

```
cloudflared-setup-cli/
├── main.go
├── go.mod
├── go.sum
├── .goreleaser.yml
├── cmd/
│   └── root.go                  # init bubbletea, launch app
├── ui/
│   ├── model.go                 # root model + screen state machine
│   ├── mainmenu.go              # main menu (5 options)
│   ├── auth/
│   │   └── model.go             # authentication screens
│   ├── credentials/
│   │   └── model.go             # credential management screens
│   ├── monitoring/
│   │   └── model.go             # observability screens (live logs, metrics)
│   ├── orchestration/
│   │   └── model.go             # orchestration screens
│   └── maintenance/
│       └── model.go             # maintenance screens
├── internal/
│   ├── cloudflared/
│   │   ├── binary.go            # detect, download, verify cloudflared binary
│   │   ├── tunnel.go            # create, list, delete, run tunnel
│   │   └── config.go            # read/write ~/.cloudflared/config.yml
│   ├── credentials/
│   │   └── store.go             # read/write ~/.cloudflared/ files
│   ├── orchestration/
│   │   ├── systemd.go           # generate + install systemd .service file
│   │   ├── docker.go            # generate docker-compose.yml
│   │   ├── windows.go           # register Windows Service via sc.exe
│   │   └── kubernetes.go        # generate Kubernetes Deployment manifest
│   ├── monitoring/
│   │   ├── logs.go              # stream cloudflared stdout/stderr
│   │   ├── status.go            # tunnel up/down status
│   │   ├── metrics.go           # Cloudflare API traffic/latency/error rate
│   │   └── health.go            # HTTP ping to local endpoints
│   ├── api/
│   │   └── cloudflare.go        # Cloudflare API client wrapper
│   └── maintenance/
│       ├── update.go            # check latest version + auto-update binary
│       ├── cleanup.go           # remove unused tunnels and configs
│       ├── backup.go            # backup and restore config files
│       └── reset.go             # uninstall/reset all cloudflared setup
└── .github/
    └── workflows/
        ├── build.yml            # build + test on push to main
        └── release.yml          # GoReleaser on tag v*.*.*
```

---

## UI Design

### Navigation Model

- Every screen displays a numbered list of options
- User presses a number key then Enter to select
- `0` always returns to the previous menu (like ATM cancel)
- Destructive actions require confirmation: `Yakin? [y/N]`
- A status bar at the bottom of every screen shows: active tunnel name + cloudflared version

### Screen Layout

```
┌─────────────────────────────────────────┐
│     CLOUDFLARED SETUP CLI  v1.0.0       │
│─────────────────────────────────────────│
│  [1] Autentikasi & Setup                │
│  [2] Manajemen Kredensial               │
│  [3] Observability & Monitoring         │
│  [4] Orkestrasi                         │
│  [5] Pemeliharaan                       │
│  [0] Keluar                             │
│─────────────────────────────────────────│
│  Tunnel: my-tunnel  |  cloudflared 2024 │
└─────────────────────────────────────────┘
```

### Menu Hierarchy

```
Main Menu
├── [1] Autentikasi & Setup
│   ├── [1] Cek instalasi cloudflared
│   ├── [2] Install / download cloudflared (auto-detect OS/arch)
│   ├── [3] Login ke Cloudflare  →  cloudflared tunnel login
│   ├── [4] Verifikasi koneksi
│   └── [0] Kembali
│
├── [2] Manajemen Kredensial
│   ├── [1] Lihat tunnel tersimpan
│   ├── [2] Buat tunnel baru
│   ├── [3] Hapus tunnel  (konfirmasi)
│   ├── [4] Konfigurasi ingress rules  (hostname → localhost:PORT)
│   ├── [5] Export / import config
│   └── [0] Kembali
│
├── [3] Observability & Monitoring
│   ├── [1] Live logs  (streaming realtime, scroll dengan viewport)
│   ├── [2] Status tunnel  (up/down, koneksi aktif)
│   ├── [3] Metrics  (traffic, latency, error rate — Cloudflare API)
│   ├── [4] Health check endpoint  (HTTP ping ke localhost)
│   └── [0] Kembali
│
├── [4] Orkestrasi
│   ├── [1] systemd service  (Linux)
│   ├── [2] Docker / Docker Compose
│   ├── [3] Windows Service
│   ├── [4] Kubernetes manifest
│   └── [0] Kembali
│
└── [5] Pemeliharaan
    ├── [1] Update cloudflared  (cek versi + auto-update)
    ├── [2] Cleanup  (hapus tunnel & config tidak terpakai)
    ├── [3] Backup & Restore config
    ├── [4] Reset / Uninstall semua  (konfirmasi)
    └── [0] Kembali
```

---

## Architecture & Data Flow

### Integration Layers

```
┌─────────────────────────────────────────────────────┐
│                cloudflared-setup-cli                 │
│                                                      │
│  UI Layer (Bubbletea)                                │
│       │                                              │
│  internal/cloudflared  ──exec──►  cloudflared binary │
│  internal/api          ──HTTP──►  api.cloudflare.com │
│  internal/credentials  ──R/W───►  ~/.cloudflared/   │
│  internal/orchestration ──gen──►  file output        │
└─────────────────────────────────────────────────────┘
```

### Operation Details

| Operation | Implementation |
|---|---|
| `cloudflared tunnel login` | `exec.Command("cloudflared", "tunnel", "login")` — opens browser |
| `cloudflared tunnel create` | exec + parse stdout for tunnel ID |
| `cloudflared tunnel run` | exec with piped Stdout, killable from UI via context cancel |
| Live logs | Read `cmd.Stdout` in goroutine → send to Bubbletea via `tea.Cmd` channel |
| Metrics | HTTP GET `api.cloudflare.com/client/v4/accounts/{id}/tunnels` with API token |
| Health check | HTTP GET to `localhost:PORT` every N seconds |
| Config files | Read/write `~/.cloudflared/config.yml` and `~/.cloudflared/cert.pem` |
| Orchestration | Generate string from template → write file → exec install command |

### Authentication Flow (Pre-requisite)

```
1. Is cloudflared binary in PATH?
   └── No  → offer auto-download (menu 1.2)
   └── Yes ↓
2. Does ~/.cloudflared/cert.pem exist?
   └── No  → run cloudflared tunnel login (menu 1.3)
   └── Yes ↓
3. Is Cloudflare API token in ~/.cloudflared/config.yml?
   └── No  → prompt for token (required for Metrics feature only)
   └── Yes → all features unlocked
```

### Error Handling

| Scenario | Behaviour |
|---|---|
| Binary not found | Show warning, redirect to menu [1] |
| Cloudflare API 401/403 | Show error message + prompt to re-login |
| Tunnel run failure | Display stderr in scrollable viewport |
| Unsupported OS for orchestration | Show "Not supported on this platform" and skip option |

---

## GitHub Workflows

### `build.yml` — on push to `main` / pull request

```yaml
matrix:
  - linux/amd64
  - linux/arm64
  - windows/amd64
  - darwin/amd64
  - darwin/arm64

steps:
  - go test ./internal/... -coverprofile=coverage.out
  - go build ./...
```

### `release.yml` — on tag `v*.*.*`

Uses `goreleaser/goreleaser-action` to:
- Build binaries for all 5 platforms
- Package as `.tar.gz` (Linux/macOS) and `.zip` (Windows)
- Generate `checksums.txt`
- Auto-generate changelog from commit messages
- Publish to GitHub Releases

### `.goreleaser.yml` artifacts

```
cloudflared-setup-cli_linux_amd64.tar.gz
cloudflared-setup-cli_linux_arm64.tar.gz
cloudflared-setup-cli_windows_amd64.zip
cloudflared-setup-cli_darwin_amd64.tar.gz
cloudflared-setup-cli_darwin_arm64.tar.gz
checksums.txt
```

---

## Installation

```bash
# Option 1 — go install (for developers)
go install github.com/adityarizkyramadhan/cloudflared-setup-cli@latest

# Option 2 — download binary from GitHub Releases
# https://github.com/adityarizkyramadhan/cloudflared-setup-cli/releases

# Option 3 — curl one-liner (Linux/macOS)
curl -fsSL https://github.com/adityarizkyramadhan/cloudflared-setup-cli/releases/latest/download/install.sh | sh
```

Versioning follows semver (`v1.0.0`). Pushing a git tag triggers the release workflow automatically.

---

## Testing Strategy

### Unit Tests (`internal/` layer)

| Package | Approach |
|---|---|
| `internal/cloudflared` | Mock `exec.Command` via interface injection |
| `internal/credentials` | Use temp directory as fake `~/.cloudflared/` |
| `internal/orchestration` | Assert generated output strings |
| `internal/monitoring/health` | Mock HTTP server on random port |

### Integration Tests

Run only when `CF_API_TOKEN` environment variable is set (via GitHub Actions secret):
- `internal/api/` — hit Cloudflare API with test account

### UI Tests

Bubbletea models are not unit tested directly. Each sub-model's `View() string` is assertable for output content. Manual smoke test via `go run .` is required before each release tag.

### Coverage Target

- `internal/` packages: ≥ 70%
- `ui/` packages: excluded from coverage requirement

---

## Bubbletea Architecture Note

The root `ui/model.go` holds a `currentScreen` enum and delegates `Init()`, `Update()`, and `View()` to the active sub-model. Screen transitions happen by returning a new model from `Update()`. Live log streaming uses `tea.Cmd` returning `tea.Msg` from a goroutine reading the cloudflared process stdout — this keeps the event loop non-blocking.

---

## Out of Scope (v1)

- Multi-account Cloudflare switching
- Web dashboard / HTTP UI
- Plugin system
- Cloudflare Access / Zero Trust policy management
