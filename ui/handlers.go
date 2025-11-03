package ui

import (
	"bubble-jira/config"
	"strings"

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
			"Done",
		},
		currentStage: 0,
		status:       "Connecting to Jira...",
		width:        terminalWidth,
	}
}

// newConfigListEditor creates a new config list editor
func newConfigListEditor(cfg *config.Config) configListEditor {
	fields := []configField{
		{key: "base_url", displayKey: Strings["SettingsConfigBaseUrl"], value: cfg.BaseURL},
        {key: "email", displayKey: Strings["SettingsConfigEmail"], value: cfg.Email},
        {key: "api_token", displayKey: Strings["SettingsConfigApiToken"], value: cfg.APIToken},
        {key: "jql", displayKey: Strings["SettingsConfigJQL"], value: cfg.JQL},
		{key: "SettingsGitIssueKeyLocation", displayKey: Strings["SettingsConfigGitIssueKeyLocation"], value: keyLeftRightMap[cfg.IssueKeyLoc]},
        {key: "SettingsGitIssueKeyStyle", displayKey: Strings["SettingsConfigGitIssueKeyStyle"], value: keyWrapperMap[cfg.IssueKeyStyle]},
		{key: "language", displayKey: Strings["language"], value: cfg.Lang},
	}

	items := make([]list.Item, len(fields))
	for i := range fields {
		items[i] = fields[i]
	}

	l := list.New(items, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	l.Title = Strings["ConfigurationTitle"]
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return configListEditor{
		fields: fields,
		list:   l,
	}
}

// newKeybindingEditor creates a new config list editor
func newKeybindingEditor(cfg *config.Config) configListEditor {
	exitKeysStr := ""
	if len(cfg.Keys.Exit) > 0 {
		exitKeysStr = strings.Join(cfg.Keys.Exit, "/")
	}

	fields := []configField{
        {key: "key_up", displayKey: Strings["SettingsKeyBindingKeyUp"], value: cfg.Keys.Up},
        {key: "key_down", displayKey: Strings["SettingsKeyBindingKeyDown"], value: cfg.Keys.Down},
        {key: "key_fast_up", displayKey: Strings["SettingsKeyBindingKeyFastUp"], value: cfg.Keys.FastUp},
        {key: "key_fast_down", displayKey: Strings["SettingsKeyBindingKeyFastDown"], value: cfg.Keys.FastDown},
        {key: "enter", displayKey: Strings["SettingsKeyBindingKeyEnter"], value: cfg.Keys.Confirm},
        {key: "back_cancel", displayKey: Strings["SettingsKeyBindingKeyBack"], value: exitKeysStr},
        {key: "add_comment", displayKey: Strings["SettingsKeyBindingComment"], value: cfg.Keys.Comment},
        {key: "search", displayKey: Strings["SettingsKeyBindingSearch"], value: cfg.Keys.Search},
    }

	items := make([]list.Item, len(fields))
	for i := range fields {
		items[i] = fields[i]
	}

	l := list.New(items, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	l.Title = Strings["ConfigurationTitle"]
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return configListEditor{
		fields: fields,
		list:   l,
	}
}

// newConfigInputEditor creates a new config input editor
func newConfigInputEditor(displayKey, value string) configInputEditor {
	t := textinput.New()
	t.Cursor.Style = focusedStyle
	t.CharLimit = 256
	t.Width = 50
	t.SetValue(value)
	t.Focus()
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	t.Placeholder = displayKey

	return configInputEditor{
		input:     t,
		key:       displayKey,
		focusSave: false,
	}
}

// updateConfigFields updates the config from the configList fields
func (m *model) updateConfigFields() {
	for _, field := range m.configList.fields {
        switch field.key {
        case "key_up":
            m.cfg.Keys.Up = field.value
        case "key_down":
            m.cfg.Keys.Down = field.value
        case "key_fast_up":
            m.cfg.Keys.FastUp = field.value
        case "key_fast_down":
            m.cfg.Keys.FastDown = field.value
        case "enter":
            m.cfg.Keys.Confirm = field.value
        case "back_cancel":
            m.cfg.Keys.Exit = strings.Split(field.value, "/")
        case "add_comment":
            m.cfg.Keys.Comment = field.value
        case "search":
            m.cfg.Keys.Search = field.value

        // Add other config fields here if needed
        case "base_url":
            m.cfg.BaseURL = field.value
        case "email":
            m.cfg.Email = field.value
        case "api_token":
            m.cfg.APIToken = field.value
        case "jql":
            m.cfg.JQL = field.value
        case "language":
            m.cfg.Lang = field.value
        }
    }
	if err := config.Save(m.cfg); err != nil {
		m.statusBar = "Error saving config: " + err.Error()
	}
}