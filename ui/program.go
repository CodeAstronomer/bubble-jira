package ui

import (
	"bubble-jira/config"
	"bubble-jira/jira"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	cfg = config.DefaultConfig()
	exitKeys, keyExitKeysStr, keyUp, keyDown, keyFastUp, keyFastDown, keyEnter = cfg.GetKeys()

	keyMap = map[string]string{
        keyUp:            "↑",
        keyDown:          "↓",
        keyFastDown:      "pgDown",
        keyFastUp:        "pgUp",
        keyEnter:         "⏎",
    }
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