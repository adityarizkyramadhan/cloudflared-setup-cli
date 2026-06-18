# cloudflared-setup-cli

ATM-style interactive TUI for managing Cloudflare Tunnels.

Runs natively on **Windows, Linux, and macOS** — the CLI auto-detects the host
OS and adapts the install location and service manager automatically.

## Install

```bash
go install github.com/adityarizkyramadhan/cloudflared-setup-cli@latest
```

Or download a binary from [Releases](https://github.com/adityarizkyramadhan/cloudflared-setup-cli/releases).

## Usage

```bash
cloudflared-setup-cli
```

Navigate with number keys. Press `0` to go back. Press `Ctrl+C` to quit.

## Menus

1. **Autentikasi & Setup** — install cloudflared (auto-detects install dir per OS), login to Cloudflare
2. **Manajemen Kredensial** — create/delete tunnels, configure ingress
3. **Observability & Monitoring** — live logs, status, health check
4. **Orkestrasi** — install as a native service auto-detected per OS (Windows Service / systemd, with UAC auto-elevation on Windows), or generate Docker Compose / Kubernetes manifests
5. **Pemeliharaan** — update, cleanup, backup, reset

## Release

Tag a commit to trigger automatic multi-platform release:

```bash
git tag v1.0.0 && git push origin v1.0.0
```
