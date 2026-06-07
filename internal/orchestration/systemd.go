package orchestration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const systemdTmpl = `[Unit]
Description=Cloudflare Tunnel — {{.TunnelName}}
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/cloudflared tunnel run {{.TunnelName}}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

func SystemdUnit(tunnelName string) (string, error) {
	tmpl, err := template.New("systemd").Parse(systemdTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"TunnelName": tunnelName}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func InstallSystemd(tunnelName string) (string, error) {
	content, err := SystemdUnit(tunnelName)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/etc/systemd/system/cloudflared-%s.service", tunnelName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write service file: %w", err)
	}
	cmds := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", fmt.Sprintf("cloudflared-%s", tunnelName)},
		{"systemctl", "start", fmt.Sprintf("cloudflared-%s", tunnelName)},
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return path, fmt.Errorf("systemctl %v: %w — %s", args[1:], err, string(out))
		}
	}
	return path, nil
}
