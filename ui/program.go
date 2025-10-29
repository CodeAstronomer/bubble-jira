package ui

import (
	"bubble-jira/jira"
	"bubble-jira/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Program wraps the bubbletea program
type Program struct {
	p *tea.Program
}

// NewProgram creates a new Bubble Tea program
func NewProgram(cfg *config.Config, jc *jira.Client) *Program {
	m := newModel(cfg, jc)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return &Program{p: p}
}

// Start starts the program
func (pr *Program) Start() error {
	return pr.p.Start()
}

// Quit sends a quit message to the program
func (pr *Program) Quit() {
	pr.p.Send(tea.Quit)
}