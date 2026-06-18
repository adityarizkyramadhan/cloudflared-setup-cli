package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/platform"
)

type orchMsg struct {
	text  string
	isErr bool
}

type orchInputState int

const (
	orchIdle orchInputState = iota
	orchWaitingTunnelName
	orchWaitingToken
)

type orchAction int

const (
	orchActionNative orchAction = iota
	orchActionDocker
	orchActionKubernetes
)

type orchestrationModel struct {
	status        string
	isErr         bool
	inputState    orchInputState
	input         string
	pendingName   string
	pendingAction orchAction
}

func newOrchestrationModel() orchestrationModel { return orchestrationModel{} }

func (m orchestrationModel) Init() tea.Cmd { return nil }

func (m orchestrationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputState != orchIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc":
				m.inputState = orchIdle
				m.input = ""
				m.status = "Dibatalkan"
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "1":
			m.pendingAction = orchActionNative
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = fmt.Sprintf("Nama tunnel untuk service native (%s): ", platform.ServiceManager())
		case "2":
			m.pendingAction = orchActionDocker
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Docker Compose: "
		case "3":
			m.pendingAction = orchActionKubernetes
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Kubernetes: "
		case "0":
			return m, GoBack()
		}
	case orchMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = orchIdle
		m.input = ""
	}
	return m, nil
}

func (m orchestrationModel) handleInput() (orchestrationModel, tea.Cmd) {
	val := strings.TrimSpace(m.input)
	switch m.inputState {
	case orchWaitingTunnelName:
		m.pendingName = val
		m.input = ""
		if m.pendingAction == orchActionDocker {
			m.inputState = orchWaitingToken
			m.status = "Tunnel token (dari Cloudflare dashboard): "
			return m, nil
		}
		m.inputState = orchIdle
		name := m.pendingName
		action := m.pendingAction
		return m, func() tea.Msg { return executeOrch(action, name, "") }
	case orchWaitingToken:
		token := val
		name := m.pendingName
		m.inputState = orchIdle
		m.input = ""
		return m, func() tea.Msg { return executeOrch(orchActionDocker, name, token) }
	}
	m.inputState = orchIdle
	return m, nil
}

func executeOrch(action orchAction, tunnelName, token string) tea.Msg {
	switch action {
	case orchActionNative:
		return installNativeService(tunnelName)
	case orchActionDocker:
		content, err := orchestration.DockerCompose(tunnelName, token)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		if err := os.WriteFile("docker-compose.yml", []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: "docker-compose.yml disimpan — jalankan: docker compose up -d"}
	case orchActionKubernetes:
		content, err := orchestration.KubernetesManifest(tunnelName)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		path := fmt.Sprintf("cloudflared-%s-deployment.yaml", tunnelName)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("Manifest disimpan: %s\nJalankan: kubectl apply -f %s", path, path)}
	}
	return orchMsg{text: "Unknown action", isErr: true}
}

// installNativeService auto-detects the host service manager and installs the
// tunnel as a native service, elevating to admin on Windows when needed.
func installNativeService(tunnelName string) tea.Msg {
	switch platform.ServiceManager() {
	case "windows":
		if !platform.IsAdmin() {
			if err := platform.RelaunchElevated(); err != nil {
				return orchMsg{text: "Butuh hak admin, tapi elevasi dibatalkan: " + err.Error(), isErr: true}
			}
			// Elevated copy takes over; this instance exits.
			return tea.QuitMsg{}
		}
		cfPath, err := exec.LookPath("cloudflared")
		if err != nil {
			dir, derr := platform.InstallDir()
			if derr != nil {
				return orchMsg{text: "cloudflared tidak ditemukan di PATH — install dulu lewat menu Autentikasi", isErr: true}
			}
			cfPath = filepath.Join(dir, "cloudflared.exe")
		}
		if err := orchestration.InstallWindowsService(tunnelName, cfPath); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("Windows Service cloudflared-%s terpasang & berjalan", tunnelName)}
	case "systemd":
		path, err := orchestration.InstallSystemd(tunnelName)
		if err != nil {
			return orchMsg{text: "Gagal pasang systemd (perlu root? jalankan ulang dengan sudo): " + err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("systemd service terpasang & aktif: %s", path)}
	case "launchd":
		return orchMsg{text: "macOS (launchd) belum didukung — gunakan Docker [2]", isErr: true}
	default:
		return orchMsg{text: "OS ini tidak didukung untuk service native — gunakan Docker [2]", isErr: true}
	}
}

func (m orchestrationModel) View() string {
	title := TitleStyle.Render("ORKESTRASI")
	menu := MenuStyle.Render(
		fmt.Sprintf("[1] Install service native (auto-detect: %s)\n", platform.ServiceManager()) +
			"[2] Docker / Docker Compose\n" +
			"[3] Kubernetes manifest\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.inputState != orchIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}
