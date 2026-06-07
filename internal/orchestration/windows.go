package orchestration

import (
	"fmt"
	"os/exec"
)

func WindowsServiceName(tunnelName string) string {
	return "cloudflared-" + tunnelName
}

func InstallWindowsService(tunnelName, cloudflaredPath string) error {
	svcName := WindowsServiceName(tunnelName)
	binPath := fmt.Sprintf(`"%s" tunnel run %s`, cloudflaredPath, tunnelName)

	cmds := [][]string{
		{"sc", "create", svcName, "binPath=", binPath, "start=", "auto"},
		{"sc", "description", svcName, "Cloudflare Tunnel — " + tunnelName},
		{"sc", "start", svcName},
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc %v: %w — %s", args[1:], err, string(out))
		}
	}
	return nil
}

func RemoveWindowsService(tunnelName string) error {
	svcName := WindowsServiceName(tunnelName)
	cmds := [][]string{
		{"sc", "stop", svcName},
		{"sc", "delete", svcName},
	}
	for _, args := range cmds {
		exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	return nil
}
