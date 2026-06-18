package cloudflared

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type TunnelInfo struct {
	ID   string
	Name string
}

func CreateTunnel(name string) (string, error) {
	out, err := exec.Command("cloudflared", "tunnel", "create", "--output", "json", name).Output()
	if err != nil {
		return "", fmt.Errorf("tunnel create: %w", err)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, l := range lines {
			if strings.Contains(l, "Created tunnel") {
				parts := strings.Fields(l)
				if len(parts) > 0 {
					return parts[len(parts)-1], nil
				}
			}
		}
		return "", fmt.Errorf("parse tunnel ID from output: %s", string(out))
	}
	return result.ID, nil
}

func ListTunnels() ([]TunnelInfo, error) {
	out, err := exec.Command("cloudflared", "tunnel", "list", "--output", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("tunnel list: %w", err)
	}
	var raw []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse tunnel list: %w", err)
	}
	tunnels := make([]TunnelInfo, len(raw))
	for i, r := range raw {
		tunnels[i] = TunnelInfo{ID: r.ID, Name: r.Name}
	}
	return tunnels, nil
}

func DeleteTunnel(name string) error {
	out, err := exec.Command("cloudflared", "tunnel", "delete", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tunnel delete: %w — %s", err, string(out))
	}
	return nil
}

// CleanupConnections disconnects stale connections for a tunnel via
// `cloudflared tunnel cleanup <name>`, freeing it for deletion.
func CleanupConnections(name string) error {
	out, err := exec.Command("cloudflared", "tunnel", "cleanup", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tunnel cleanup: %w — %s", err, string(out))
	}
	return nil
}

// isActiveConnectionsErr reports whether a delete failure is the "tunnel has
// active connections" case (Cloudflare error code 1022).
func isActiveConnectionsErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "active connection") || strings.Contains(s, "1022")
}

// DeleteTunnelWithCleanup deletes a tunnel. If the tunnel reports active
// connections, it disconnects stale connections and retries once. A tunnel
// that is still actively running (e.g. via a service) returns a clear error
// telling the user to stop it first.
func DeleteTunnelWithCleanup(name string) error {
	err := DeleteTunnel(name)
	if err == nil {
		return nil
	}
	if !isActiveConnectionsErr(err) {
		return err
	}
	if cerr := CleanupConnections(name); cerr != nil {
		return fmt.Errorf("%w; cleanup koneksi juga gagal: %v", err, cerr)
	}
	if err2 := DeleteTunnel(name); err2 != nil {
		if isActiveConnectionsErr(err2) {
			return fmt.Errorf("tunnel %q masih punya koneksi aktif — stop service/proses cloudflared yang menjalankannya dulu, lalu hapus lagi", name)
		}
		return err2
	}
	return nil
}

func RunTunnel(name string, w io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command("cloudflared", "tunnel", "run", name)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("tunnel run: %w", err)
	}
	return cmd, nil
}

func RouteDNS(tunnelName, hostname string) error {
	out, err := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelName, hostname).CombinedOutput()
	if err != nil {
		return fmt.Errorf("route dns: %w — %s", err, string(out))
	}
	return nil
}

func CleanupTunnel(names []string) []error {
	var errs []error
	for _, name := range names {
		if err := DeleteTunnel(name); err != nil {
			errs = append(errs, fmt.Errorf("delete %q: %w", name, err))
		}
	}
	return errs
}

func StopTunnel(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(os.Interrupt)
}
