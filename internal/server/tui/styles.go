package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	yellowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	redStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
)
