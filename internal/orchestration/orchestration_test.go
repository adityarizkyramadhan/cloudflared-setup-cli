package orchestration_test

import (
	"strings"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
)

func TestSystemdUnit(t *testing.T) {
	out, err := orchestration.SystemdUnit("my-tunnel")
	if err != nil {
		t.Fatalf("SystemdUnit: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in systemd unit")
	}
	if !strings.Contains(out, "ExecStart=") {
		t.Error("expected ExecStart in systemd unit")
	}
}

func TestDockerCompose(t *testing.T) {
	out, err := orchestration.DockerCompose("my-tunnel", "tok123")
	if err != nil {
		t.Fatalf("DockerCompose: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in docker-compose")
	}
	if !strings.Contains(out, "tok123") {
		t.Error("expected token in docker-compose")
	}
}

func TestKubernetesManifest(t *testing.T) {
	out, err := orchestration.KubernetesManifest("my-tunnel")
	if err != nil {
		t.Fatalf("KubernetesManifest: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in k8s manifest")
	}
	if !strings.Contains(out, "kind: Deployment") {
		t.Error("expected Deployment kind")
	}
}

func TestWindowsServiceName(t *testing.T) {
	name := orchestration.WindowsServiceName("my-tunnel")
	if name != "cloudflared-my-tunnel" {
		t.Errorf("got %q", name)
	}
}
