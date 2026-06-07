package orchestration

import (
	"bytes"
	"text/template"
)

const dockerComposeTmpl = `version: "3.8"

services:
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel run {{.TunnelName}}
    environment:
      - TUNNEL_TOKEN={{.TunnelToken}}
    volumes:
      - ~/.cloudflared:/etc/cloudflared
`

func DockerCompose(tunnelName, tunnelToken string) (string, error) {
	tmpl, err := template.New("docker").Parse(dockerComposeTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"TunnelName":  tunnelName,
		"TunnelToken": tunnelToken,
	})
	return buf.String(), err
}
