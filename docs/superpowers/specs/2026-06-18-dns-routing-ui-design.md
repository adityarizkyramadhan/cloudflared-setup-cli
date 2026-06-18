# DNS Routing in the UI — Design

**Date:** 2026-06-18
**Status:** Approved

## Problem

The backend already exposes `RouteDNS` (`internal/cloudflared/tunnel.go`, which
runs `cloudflared tunnel route dns <tunnel> <hostname>`), but no UI screen calls
it. As a result a user can create a tunnel and add ingress rules, yet the
hostname never gets a DNS CNAME pointing at the tunnel — so traffic does not
resolve. DNS routing is the missing step in the tunnel setup lifecycle.

## Goal

Wire DNS routing into the TUI so the hostname → tunnel CNAME is created as part
of the normal flow, with a manual option for re-routing.

## Decisions (from brainstorming)

1. Expose DNS routing **both** integrated into the ingress flow **and** as a
   standalone menu item.
2. The tunnel name comes from `config.yml`; if absent, prompt the user.

## Non-Goals (YAGNI)

- Editing or deleting ingress rules.
- Hostname validation.
- Managing multiple active tunnels in config.yml (single active tunnel only).

## Architecture

### Backend: `internal/cloudflared/config.go` (2 new helpers)

- `ActiveTunnel() (string, error)` — reads `config.yml` and returns the `tunnel`
  field. Returns `("", nil)` when the file or field is absent (not an error).
- `SetTunnel(name string) error` — reads the config, sets `Tunnel`, writes back
  (reusing existing `ReadConfig`/`WriteConfig`).

`RouteDNS(tunnelName, hostname string) error` already exists and is reused
unchanged.

### UI: `ui/credentials.go`

New model fields: `pendingHostname`, `pendingService`.
New input states: `credWaitingIngressTunnel`, `credWaitingDNSHostname`,
`credWaitingDNSTunnel`.

**Create tunnel `[2]`** — after `CreateTunnel(name)` succeeds, call
`SetTunnel(name)` so `config.yml` records the active tunnel. Success message
notes the tunnel was set active. A `SetTunnel` failure is reported but does not
undo the created tunnel.

**Ingress flow `[4]`** — after collecting hostname + service:
1. Read `ActiveTunnel()`.
2. If non-empty → dispatch a single command that runs `AddIngressRule` then
   `RouteDNS`.
3. If empty → prompt `credWaitingIngressTunnel`
   ("Nama tunnel untuk DNS route (Enter kosong = skip): "), then dispatch the
   same combined command (skipping DNS if the user left it blank).

**Menu `[6] Route DNS`** — standalone:
1. Prompt hostname (`credWaitingDNSHostname`).
2. Resolve tunnel from `ActiveTunnel()`; if empty, prompt
   (`credWaitingDNSTunnel`).
3. Run `RouteDNS(tunnel, hostname)`.

## Data Flow (ingress + DNS)

```
[4] → hostname → service
   → ActiveTunnel()
       non-empty → AddIngressRule + RouteDNS → combined message
       empty     → prompt tunnel
                     blank → AddIngressRule only (DNS skipped)
                     name  → AddIngressRule + RouteDNS → combined message
```

## Error Handling

- `AddIngressRule` failure is fatal to the action (reported as error, no DNS
  attempted).
- `RouteDNS` failure is **non-fatal**: the ingress rule is already saved, so the
  result shows the ingress success plus a ⚠ warning with the DNS error (e.g.
  "record already exists", "not logged in"). The combined message uses the
  success style with an inline warning, not the error style.
- Standalone `[6]` `RouteDNS` failure is reported as an error (nothing else to
  preserve).

## Testing

- `internal/cloudflared/config_test.go`:
  - `TestSetTunnel` — `SetTunnel` then `ReadConfig` shows the tunnel.
  - `TestActiveTunnel` — returns the written tunnel; returns `""` with no error
    when config is absent. Uses the existing `t.Setenv("HOME"/"USERPROFILE")`
    temp-dir pattern.
- `RouteDNS` (shells out to `cloudflared`) and the TUI flow are verified
  manually, consistent with existing exec/UI code that is not unit-tested.
- Existing tests stay green; `go build`/`go vet` clean on all platforms.
