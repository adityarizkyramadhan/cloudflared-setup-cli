package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	MenuStyle = lipgloss.NewStyle().Padding(0, 2)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
