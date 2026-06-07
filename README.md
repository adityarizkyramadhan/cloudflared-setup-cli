# cloudflared-setup-cli

ATM-style interactive TUI for managing Cloudflare Tunnels.

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

1. **Autentikasi & Setup** — install cloudflared, login to Cloudflare
2. **Manajemen Kredensial** — create/delete tunnels, configure ingress
3. **Observability & Monitoring** — live logs, status, health check
4. **Orkestrasi** — generate systemd, Docker Compose, Windows Service, or Kubernetes manifests
5. **Pemeliharaan** — update, cleanup, backup, reset

## Release

Tag a commit to trigger automatic multi-platform release:

```bash
git tag v1.0.0 && git push origin v1.0.0
```
