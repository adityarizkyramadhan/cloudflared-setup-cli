package orchestration

import (
	"bytes"
	"text/template"
)

const kubernetesTmpl = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared-{{.TunnelName}}
  labels:
    app: cloudflared
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      containers:
        - name: cloudflared
          image: cloudflare/cloudflared:latest
          args:
            - tunnel
            - run
            - {{.TunnelName}}
          env:
            - name: TUNNEL_TOKEN
              valueFrom:
                secretKeyRef:
                  name: cloudflared-{{.TunnelName}}
                  key: token
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 128Mi
`

func KubernetesManifest(tunnelName string) (string, error) {
	tmpl, err := template.New("k8s").Parse(kubernetesTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{"TunnelName": tunnelName})
	return buf.String(), err
}
