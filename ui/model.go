package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifies which sub-model is active.
type Screen int

const (
	ScreenMain Screen = iota
	ScreenAuth
	ScreenCredentials
	ScreenMonitoring
	ScreenOrchestration
	ScreenMaintenance
)

// NavigateMsg tells the root model to switch to a new screen.
type NavigateMsg struct{ To Screen }

// GoBackMsg returns to the main menu.
type GoBackMsg struct{}

// NavigateTo returns a Cmd that sends a NavigateMsg.
func NavigateTo(s Screen) tea.Cmd {
	return func() tea.Msg { return NavigateMsg{To: s} }
}

// GoBack returns a Cmd that sends a GoBackMsg.
func GoBack() tea.Cmd {
	return func() tea.Msg { return GoBackMsg{} }
}

// RootModel is the top-level Bubbletea model.
type RootModel struct {
	screen  Screen
	current tea.Model
}

func NewRootModel() RootModel {
	return RootModel{
		screen:  ScreenMain,
		current: newMainMenuModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case NavigateMsg:
		m.screen = msg.To
		m.current = screenFor(msg.To)
		return m, m.current.Init()
	case GoBackMsg:
		m.screen = ScreenMain
		m.current = newMainMenuModel()
		return m, m.current.Init()
	}

	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	return m.current.View()
}

// screenFor returns a fresh model for the given screen.
// Sub-models are wired in Task 15; stubs return main menu until then.
func screenFor(s Screen) tea.Model {
	return newMainMenuModel() // replaced in Task 15
}
