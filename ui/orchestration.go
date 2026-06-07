package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
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
	orchActionSystemd orchAction = iota
	orchActionDocker
	orchActionWindows
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
			m.pendingAction = orchActionSystemd
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk systemd service: "
		case "2":
			m.pendingAction = orchActionDocker
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Docker Compose: "
		case "3":
			m.pendingAction = orchActionWindows
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Windows Service: "
		case "4":
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
	case orchActionSystemd:
		content, err := orchestration.SystemdUnit(tunnelName)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		path := fmt.Sprintf("cloudflared-%s.service", tunnelName)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("Service file disimpan: %s\nJalankan: sudo cp %s /etc/systemd/system/ && sudo systemctl enable --now %s", path, path, path[:len(path)-8])}
	case orchActionDocker:
		content, err := orchestration.DockerCompose(tunnelName, token)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		if err := os.WriteFile("docker-compose.yml", []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: "docker-compose.yml disimpan — jalankan: docker compose up -d"}
	case orchActionWindows:
		return orchMsg{text: fmt.Sprintf("Jalankan sebagai Admin:\nsc create cloudflared-%s binPath=\"cloudflared.exe tunnel run %s\" start=auto\nsc start cloudflared-%s", tunnelName, tunnelName, tunnelName)}
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

func (m orchestrationModel) View() string {
	title := TitleStyle.Render("ORKESTRASI")
	menu := MenuStyle.Render(
		"[1] systemd service  (Linux)\n" +
			"[2] Docker / Docker Compose\n" +
			"[3] Windows Service\n" +
			"[4] Kubernetes manifest\n\n" +
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
