package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Alice/pain_tz/internal/server"
)

// refreshMsg is sent periodically to refresh the table.
type refreshMsg struct{}

// Model is the bubbletea model for the top-like TUI.
type Model struct {
	store        *server.Store
	table        table.Model
	staleTimeout time.Duration
	width        int
	height       int
	quitting     bool
}

// NewModel creates a new TUI model backed by the given store.
func NewModel(store *server.Store, staleTimeout time.Duration) Model {
	columns := []table.Column{
		{Title: "Host", Width: 20},
		{Title: "CPU%", Width: 8},
		{Title: "Mem%", Width: 8},
		{Title: "Disk%", Width: 8},
		{Title: "Net↑", Width: 10},
		{Title: "Net↓", Width: 10},
		{Title: "Seen", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = headerStyle
	s.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED"))
	t.SetStyles(s)

	return Model{
		store:        store,
		table:        t,
		staleTimeout: staleTimeout,
	}
}

// Init returns the initial command (tick at 1 second).
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

// Update handles messages (tick, key, window resize).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case refreshMsg:
		m.refreshRows()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return refreshMsg{}
		})

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5) // leave room for header + status + help
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return "bye.\n"
	}

	title := titleStyle.Render(fmt.Sprintf(" pain_tz · %d hosts monitored ", m.store.HostCount()))
	status := m.renderStatus()
	help := helpStyle.Render("q: quit  ↑↓: scroll  ·  refreshed every 1s")
	table := m.table.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		table,
		status,
		help,
	)
}

// refreshRows rebuilds table rows from the store.
func (m *Model) refreshRows() {
	hosts := m.store.GetAll()
	now := time.Now()
	rows := make([]table.Row, len(hosts))

	for i, h := range hosts {
		r := h.Report
		age := now.Sub(h.LastSeen).Seconds()
		stale := age > m.staleTimeout.Seconds()

		cpuPct := r.Cpu.UsagePercent
		memPct := r.Memory.UsedPercent

		// Worst disk usage
		diskPct := 0.0
		for _, d := range r.Disks {
			if d.UsedPercent > diskPct {
				diskPct = d.UsedPercent
			}
		}

		// Net totals
		var netUp, netDown uint64
		for _, n := range r.Networks {
			netUp += n.BytesSent
			netDown += n.BytesRecv
		}

		seen := "online"
		if stale {
			seen = fmt.Sprintf("stale %.0fs", age)
		} else {
			seen = fmt.Sprintf("%.0fs ago", age)
		}

		// Colorize based on thresholds
		styleCPU := lipgloss.NewStyle().Foreground(colorForPercent(cpuPct))
		styleMem := lipgloss.NewStyle().Foreground(colorForPercent(memPct))
		styleDisk := lipgloss.NewStyle().Foreground(colorForPercent(diskPct))
		styleSeen := lipgloss.NewStyle().Foreground(colorForStatus(!stale, age))

		rows[i] = table.Row{
			h.HostID,
			styleCPU.Render(fmt.Sprintf("%.1f", cpuPct)),
			styleMem.Render(fmt.Sprintf("%.1f", memPct)),
			styleDisk.Render(fmt.Sprintf("%.1f", diskPct)),
			formatBytes(netUp),
			formatBytes(netDown),
			styleSeen.Render(seen),
		}
	}

	m.table.SetRows(rows)
}

func (m Model) renderStatus() string {
	hosts := m.store.GetAll()
	online := 0
	stale := 0
	now := time.Now()
	for _, h := range hosts {
		if now.Sub(h.LastSeen).Seconds() > m.staleTimeout.Seconds() {
			stale++
		} else {
			online++
		}
	}
	return statusBarStyle.Render(fmt.Sprintf("online: %d  stale: %d  total: %d", online, stale, len(hosts)))
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
