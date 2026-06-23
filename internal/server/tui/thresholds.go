package tui

import "github.com/charmbracelet/lipgloss"

// Threshold colors for resource utilization.
func colorForPercent(pct float64) lipgloss.Color {
	switch {
	case pct >= 80:
		return lipgloss.Color("#FF0000") // red
	case pct >= 50:
		return lipgloss.Color("#FFAA00") // yellow
	default:
		return lipgloss.Color("#00FF00") // green
	}
}

func colorForStatus(online bool, lastSeenSeconds float64) lipgloss.Color {
	if !online {
		return lipgloss.Color("#FF0000") // red: stale
	}
	if lastSeenSeconds > 15 {
		return lipgloss.Color("#FFAA00") // yellow: slow
	}
	return lipgloss.Color("#00FF00") // green: healthy
}
