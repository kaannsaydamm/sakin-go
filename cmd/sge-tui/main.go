package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	services []string
	status   map[string]string
}

func initialModel() model {
	return model{
		services: []string{"Ingest", "Correlation", "Analytics", "Richment", "SOAR", "Panel"},
		status: map[string]string{
			"Ingest":      "Starting...",
			"Correlation": "Waiting...",
			"Analytics":   "Waiting...",
			"Richment":    "Waiting...",
			"SOAR":        "Waiting...",
			"Panel":       "Waiting...",
		},
	}
}

// TickMsg is sent to update the UI
type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case TickMsg:
		// Simulate status updates
		m.status["Ingest"] = "Running ✅"
		m.status["Correlation"] = "Running ✅"
		m.status["Analytics"] = "Processing ⚡"
		m.status["Richment"] = "Cached 💎"
		m.status["SOAR"] = "Idle 💤"
		m.status["Panel"] = "Listening 🌐"
		return m, tick()
	}
	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	rowStyle = lipgloss.NewStyle().
			PaddingLeft(2)
)

func (m model) View() string {
	s := titleStyle.Render("S.A.K.I.N. Go Edition - Terminal Dashboard") + "\n\n"

	for _, svc := range m.services {
		status := m.status[svc]
		s += rowStyle.Render(fmt.Sprintf("%-15s : %s", svc, status)) + "\n"
	}

	s += "\nPress 'q' to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
