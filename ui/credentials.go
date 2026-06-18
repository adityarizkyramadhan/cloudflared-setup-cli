package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/credentials"
)

type credMsg struct {
	text  string
	isErr bool
}

type credInputState int

const (
	credIdle credInputState = iota
	credWaitingTunnelName
	credWaitingHostname
	credWaitingService
	credWaitingDeleteName
	credWaitingExportPath
	credWaitingIngressTunnel
	credWaitingDNSHostname
	credWaitingDNSTunnel
)

type credentialsModel struct {
	status          string
	isErr           bool
	inputState      credInputState
	input           string
	pendingHostname string
	pendingService  string
}

func newCredentialsModel() credentialsModel { return credentialsModel{} }

func (m credentialsModel) Init() tea.Cmd { return nil }

func (m credentialsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputState != credIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc", "ctrl+c":
				m.inputState = credIdle
				m.input = ""
				m.pendingHostname = ""
				m.pendingService = ""
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
			return m, listTunnels
		case "2":
			m.inputState = credWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel: "
		case "3":
			m.inputState = credWaitingDeleteName
			m.input = ""
			m.status = "Nama tunnel yang akan dihapus: "
		case "4":
			m.inputState = credWaitingHostname
			m.input = ""
			m.status = "Hostname (contoh: app.domain.com): "
		case "5":
			return m, exportConfig
		case "6":
			m.inputState = credWaitingDNSHostname
			m.input = ""
			m.status = "Hostname untuk DNS route (contoh: app.domain.com): "
		case "0":
			return m, GoBack()
		}
	case credMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = credIdle
		m.input = ""
	}
	return m, nil
}

func (m credentialsModel) handleInput() (credentialsModel, tea.Cmd) {
	switch m.inputState {
	case credWaitingTunnelName:
		name := strings.TrimSpace(m.input)
		m.inputState = credIdle
		m.input = ""
		return m, func() tea.Msg {
			id, err := cloudflared.CreateTunnel(name)
			if err != nil {
				return credMsg{text: err.Error(), isErr: true}
			}
			if err := cloudflared.SetTunnel(name); err != nil {
				return credMsg{text: fmt.Sprintf("Tunnel %q dibuat (ID: %s), tapi gagal diset aktif: %v", name, id, err)}
			}
			return credMsg{text: fmt.Sprintf("Tunnel %q dibuat & diset aktif — ID: %s", name, id)}
		}
	case credWaitingDeleteName:
		name := strings.TrimSpace(m.input)
		m.inputState = credIdle
		m.input = ""
		return m, func() tea.Msg {
			if err := cloudflared.DeleteTunnelWithCleanup(name); err != nil {
				return credMsg{text: err.Error(), isErr: true}
			}
			return credMsg{text: fmt.Sprintf("Tunnel %q dihapus", name)}
		}
	case credWaitingHostname:
		m.pendingHostname = strings.TrimSpace(m.input)
		m.inputState = credWaitingService
		m.input = ""
		m.status = "Service (contoh: http://localhost:8080): "
		return m, nil
	case credWaitingService:
		m.pendingService = strings.TrimSpace(m.input)
		m.input = ""
		tunnel, _ := cloudflared.ActiveTunnel()
		if tunnel != "" {
			host, svc := m.pendingHostname, m.pendingService
			m.inputState = credIdle
			m.pendingHostname, m.pendingService = "", ""
			return m, func() tea.Msg { return addIngressAndRoute(host, svc, tunnel) }
		}
		m.inputState = credWaitingIngressTunnel
		m.status = "Nama tunnel untuk DNS route (Enter kosong = lewati DNS): "
		return m, nil
	case credWaitingIngressTunnel:
		tunnel := strings.TrimSpace(m.input)
		host, svc := m.pendingHostname, m.pendingService
		m.inputState = credIdle
		m.input = ""
		m.pendingHostname, m.pendingService = "", ""
		return m, func() tea.Msg { return addIngressAndRoute(host, svc, tunnel) }
	case credWaitingDNSHostname:
		m.pendingHostname = strings.TrimSpace(m.input)
		m.input = ""
		tunnel, _ := cloudflared.ActiveTunnel()
		if tunnel != "" {
			host := m.pendingHostname
			m.inputState = credIdle
			m.pendingHostname = ""
			return m, func() tea.Msg { return routeDNSOnly(tunnel, host) }
		}
		m.inputState = credWaitingDNSTunnel
		m.status = "Nama tunnel: "
		return m, nil
	case credWaitingDNSTunnel:
		tunnel := strings.TrimSpace(m.input)
		host := m.pendingHostname
		m.inputState = credIdle
		m.input = ""
		m.pendingHostname = ""
		if tunnel == "" {
			return m, func() tea.Msg { return credMsg{text: "Dibatalkan — nama tunnel kosong", isErr: true} }
		}
		return m, func() tea.Msg { return routeDNSOnly(tunnel, host) }
	}
	return m, nil
}

func (m credentialsModel) View() string {
	title := TitleStyle.Render("MANAJEMEN KREDENSIAL")
	menu := MenuStyle.Render(
		"[1] Lihat tunnel tersimpan\n" +
			"[2] Buat tunnel baru\n" +
			"[3] Hapus tunnel\n" +
			"[4] Konfigurasi ingress rules\n" +
			"[5] Export / import config\n" +
			"[6] Route DNS (CNAME ke tunnel)\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.inputState != credIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render("✓ "+m.status)
		}
	}
	return title + "\n" + menu + bottom
}

func listTunnels() tea.Msg {
	tunnels, err := cloudflared.ListTunnels()
	if err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	if len(tunnels) == 0 {
		return credMsg{text: "Tidak ada tunnel"}
	}
	var sb strings.Builder
	for _, t := range tunnels {
		fmt.Fprintf(&sb, "• %s (%s)\n", t.Name, t.ID)
	}
	return credMsg{text: strings.TrimSpace(sb.String())}
}

func exportConfig() tea.Msg {
	store, err := credentials.New()
	if err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	if err := store.BackupTo("./cloudflared-backup"); err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	return credMsg{text: "Config di-export ke ./cloudflared-backup"}
}

// addIngressAndRoute adds the ingress rule, then creates the DNS route for the
// hostname. A RouteDNS failure is non-fatal: the ingress rule is kept and the
// DNS error is surfaced as a warning.
func addIngressAndRoute(hostname, service, tunnel string) tea.Msg {
	if err := cloudflared.AddIngressRule(hostname, service); err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	base := fmt.Sprintf("Ingress %s → %s ditambahkan", hostname, service)
	if tunnel == "" {
		return credMsg{text: base + "\n⚠ DNS route dilewati (tunnel tidak diketahui)"}
	}
	if err := cloudflared.RouteDNS(tunnel, hostname); err != nil {
		return credMsg{text: base + fmt.Sprintf("\n⚠ DNS route gagal: %v", err)}
	}
	return credMsg{text: base + fmt.Sprintf("\nDNS route dibuat: %s → %s", hostname, tunnel)}
}

func routeDNSOnly(tunnel, hostname string) tea.Msg {
	if err := cloudflared.RouteDNS(tunnel, hostname); err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	return credMsg{text: fmt.Sprintf("DNS route dibuat: %s → %s", hostname, tunnel)}
}
