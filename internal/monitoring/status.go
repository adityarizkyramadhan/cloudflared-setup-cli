package monitoring

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type TunnelStatus struct {
	Name   string
	Status string
}

func GetStatus(tunnelName string) (*TunnelStatus, error) {
	out, err := exec.Command("cloudflared", "tunnel", "info", "--output", "json", tunnelName).Output()
	if err != nil {
		return &TunnelStatus{Name: tunnelName, Status: "unknown"}, nil
	}
	var raw struct {
		Status string `json:"status"`
		Name   string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse tunnel info: %w", err)
	}
	return &TunnelStatus{Name: raw.Name, Status: raw.Status}, nil
}

func GetAllStatuses(names []string) []TunnelStatus {
	result := make([]TunnelStatus, 0, len(names))
	for _, name := range names {
		s, err := GetStatus(name)
		if err != nil || s == nil {
			result = append(result, TunnelStatus{Name: name, Status: "unknown"})
			continue
		}
		result = append(result, *s)
	}
	return result
}
