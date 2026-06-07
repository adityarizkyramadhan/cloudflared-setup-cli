package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/monitoring"
)

type logLineMsg string
type logDoneMsg struct{}
type monitorMsg struct {
	text  string
	isErr bool
}

type monitoringSubScreen int

const (
	monitorMain monitoringSubScreen = iota
	monitorLogs
)

type monitoringModel struct {
	subScreen monitoringSubScreen
	viewport  viewport.Model
	logLines  []string
	streamer  *monitoring.LogStreamer
	status    string
	isErr     bool
	ready     bool
}

func newMonitoringModel() monitoringModel {
	vp := viewport.New(80, 20)
	return monitoringModel{viewport: vp}
}

func (m monitoringModel) Init() tea.Cmd { return nil }

func (m monitoringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		m.ready = true

	case tea.KeyMsg:
		if m.subScreen == monitorLogs {
			switch msg.String() {
			case "q", "0":
				if m.streamer != nil {
					m.streamer.Stop()
					m.streamer = nil
				}
				m.subScreen = monitorMain
				m.logLines = nil
				return m, nil
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "1":
			m.subScreen = monitorLogs
			m.logLines = nil
			return m, startLogStream("my-tunnel")
		case "2":
			return m, checkStatus
		case "3":
			return m, fetchMetrics
		case "4":
			return m, runHealthCheck
		case "0":
			return m, GoBack()
		}

	case logLineMsg:
		m.logLines = append(m.logLines, string(msg))
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))
		m.viewport.GotoBottom()
		return m, readNextLine(m.streamer)

	case logDoneMsg:
		m.logLines = append(m.logLines, "[stream ended]")
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))

	case monitorMsg:
		m.status = msg.text
		m.isErr = msg.isErr
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m monitoringModel) View() string {
	if m.subScreen == monitorLogs {
		header := TitleStyle.Render("LIVE LOGS") + DimStyle.Render("  [q] berhenti")
		return header + "\n" + m.viewport.View()
	}

	title := TitleStyle.Render("OBSERVABILITY & MONITORING")
	menu := MenuStyle.Render(
		"[1] Live logs\n" +
			"[2] Status tunnel\n" +
			"[3] Metrics (Cloudflare API)\n" +
			"[4] Health check endpoint\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}

func startLogStream(tunnelName string) tea.Cmd {
	return func() tea.Msg {
		streamer, err := monitoring.NewLogStreamer(tunnelName)
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		line, err := streamer.NextLine()
		if err == io.EOF {
			return logDoneMsg{}
		}
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		_ = streamer
		return logLineMsg(line)
	}
}

func readNextLine(streamer *monitoring.LogStreamer) tea.Cmd {
	if streamer == nil {
		return nil
	}
	return func() tea.Msg {
		line, err := streamer.NextLine()
		if err == io.EOF {
			return logDoneMsg{}
		}
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		return logLineMsg(line)
	}
}

func checkStatus() tea.Msg {
	cfg, err := readActiveTunnelName()
	if err != nil || cfg == "" {
		return monitorMsg{text: "Tidak ada tunnel aktif di config", isErr: true}
	}
	s, err := monitoring.GetStatus(cfg)
	if err != nil {
		return monitorMsg{text: err.Error(), isErr: true}
	}
	return monitorMsg{text: fmt.Sprintf("Tunnel %q: %s", s.Name, s.Status)}
}

func fetchMetrics() tea.Msg {
	return monitorMsg{text: "Metrics memerlukan CF_API_TOKEN dan ACCOUNT_ID — set di config"}
}

func runHealthCheck() tea.Msg {
	result := monitoring.CheckHealth("http://localhost:8080")
	return monitorMsg{text: monitoring.FormatHealth(result)}
}

func readActiveTunnelName() (string, error) {
	cfg, err := cloudflared.ReadConfig()
	if err != nil {
		return "", err
	}
	return cfg.Tunnel, nil
}
