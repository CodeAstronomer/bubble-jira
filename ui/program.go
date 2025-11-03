package ui

import (
    "fmt"
    "os"

	"bubble-jira/jira"
	"bubble-jira/config"

    "github.com/charmbracelet/bubbles/list"
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

func (pr *Program) StartWithTasks(cfg *config.Config, jc *jira.Client) error {
    // Check config validity first
    if !cfg.IsValid() {
        fmt.Fprintln(os.Stdout, "⚠ Config incomplete! Please fill all fields in Settings > Edit Config first.")
        os.Exit(1)
    }

	// Create model in fetching state
	m := newModel(cfg, jc)
	m.state = "fetching"
	m.fetching = newFetchingModel()
	m.configValidError = ""

	// Create the program
	pr.p = tea.NewProgram(m, tea.WithAltScreen())

	// Trigger the fetch commands immediately after start
	go func() {
		pr.p.Send(startTasksMsg{})
	}()

	// Start the program
	return pr.p.Start()
}

func (pr *Program) StartWithSettings(cfg *config.Config, jc *jira.Client) error {
	// Create model
    m := newModel(cfg, jc)
    m.state = "settings"
    m.configValidError = ""

    settingsItems := []list.Item{
        menuItem{title: Strings["MenuConfigTitle"], enabled: true},
        menuItem{title: Strings["MenuLicenceTitle"], enabled: true},
        menuItem{title: Strings["MenuKeybindingTitle"], enabled: true},
        menuItem{title: Strings["MenuBackTitle"], enabled: true},
    }
    m.settings = list.New(settingsItems, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
    m.settings.Title = Strings["SettingsTitle"]
    m.settings.SetShowHelp(true)
    m.settings.SetShowPagination(false)

	// Create the program
	pr.p = tea.NewProgram(m, tea.WithAltScreen())

	// Start the program
	return pr.p.Start()
}