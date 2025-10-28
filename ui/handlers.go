package ui

import (
	"bubble-jira/config"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// newFetchingModel creates a new fetching model
func newFetchingModel() fetchingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(terminalWidth),
		progress.WithoutPercentage(),
	)

	return fetchingModel{
		spinner:      s,
		progress:     p,
		stages: []string{
			"Connecting to Jira...",
			"Authenticating...",
			"Fetching comments...",
			"Processing results...",
		},
		currentStage: 0,
		status:       "Connecting to Jira...",
		width:        terminalWidth,
	}
}

// newConfigListEditor creates a new config list editor
func newConfigListEditor(cfg *config.Config) configListEditor {
	fields := []configField{
		{key: "Base URL", value: cfg.BaseURL},
		{key: "Email", value: cfg.Email},
		{key: "API Token", value: cfg.APIToken},
		{key: "JQL Query", value: cfg.JQL},
	}

	items := make([]list.Item, len(fields))
	for i := range fields {
		items[i] = fields[i]
	}

	l := list.New(items, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	l.Title = "Configuration"
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return configListEditor{
		fields: fields,
		list:   l,
	}
}

// newConfigInputEditor creates a new config input editor
func newConfigInputEditor(key, value string) configInputEditor {
	t := textinput.New()
	t.Cursor.Style = focusedStyle
	t.CharLimit = 256
	t.Width = 50
	t.SetValue(value)
	t.Focus()
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	t.Placeholder = key

	return configInputEditor{
		input:     t,
		key:       key,
		focusSave: false,
	}
}

// updateConfigFields updates the config from the configList fields
func (m *model) updateConfigFields() {
	for _, field := range m.configList.fields {
		switch field.key {
		case "Base URL":
			m.cfg.BaseURL = field.value
		case "Email":
			m.cfg.Email = field.value
		case "API Token":
			m.cfg.APIToken = field.value
		case "JQL Query":
			m.cfg.JQL = field.value
		}
	}
	if err := config.Save(m.cfg); err != nil {
		m.statusBar = "Error saving config: " + err.Error()
	}
}