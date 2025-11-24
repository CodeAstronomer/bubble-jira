package ui

import (
	"bubble-jira/config"
	"bubble-jira/ui/types"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// newFetchingModel creates a new fetching model
func newFetchingModel() types.FetchingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(terminalWidth),
		progress.WithoutPercentage(),
	)

	return types.FetchingModel{
		Spinner:      s,
		Progress:     p,
		Stages: []string{
			Strings["FetchStage1"],
			Strings["FetchStage2"],
			Strings["FetchStage3"],
			Strings["FetchStage4"],
			Strings["FetchStage5"],
		},
		CurrentStage: 0,
		Status:       Strings["FetchStage1"],
		Width:        terminalWidth,
	}
}

// newConfigListEditor creates a new config list editor
func newConfigListEditor(cfg *config.Config) types.ConfigListEditor {
	fields := []types.ConfigField{
		{Key: "base_url", DisplayKey: Strings["SettingsConfigBaseUrl"], Value: cfg.BaseURL},
        {Key: "email", DisplayKey: Strings["SettingsConfigEmail"], Value: cfg.Email},
        {Key: "api_token", DisplayKey: Strings["SettingsConfigApiToken"], Value: cfg.APIToken},
        {Key: "jql", DisplayKey: Strings["SettingsConfigJQL"], Value: cfg.JQL},
		{Key: "SettingsGitIssueKeyLocation", DisplayKey: Strings["SettingsConfigGitIssueKeyLocation"], Value: keyLeftRightMap[cfg.IssueKeyLoc]},
        {Key: "SettingsGitIssueKeyStyle", DisplayKey: Strings["SettingsConfigGitIssueKeyStyle"], Value: keyWrapperMap[cfg.IssueKeyStyle]},
		{Key: "SettingsLanguage", DisplayKey: Strings["language"], Value: cfg.Lang},
	}

	items := make([]list.Item, len(fields))
	for i := range fields {
		items[i] = fields[i]
	}

	l := list.New(items, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	l.Title = Strings["ConfigurationTitle"]
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return types.ConfigListEditor{
		Fields: fields,
		List:   l,
	}
}

// newKeybindingEditor creates a new config list editor
func newKeybindingEditor(cfg *config.Config) types.ConfigListEditor {
	exitKeysStr := ""
	if len(cfg.Keys.Exit) > 0 {
		exitKeysStr = strings.Join(cfg.Keys.Exit, "/")
	}

	fields := []types.ConfigField{
        {Key: "key_up", DisplayKey: Strings["SettingsKeyBindingKeyUp"], Value: cfg.Keys.Up},
        {Key: "key_down", DisplayKey: Strings["SettingsKeyBindingKeyDown"], Value: cfg.Keys.Down},
        {Key: "key_fast_up", DisplayKey: Strings["SettingsKeyBindingKeyFastUp"], Value: cfg.Keys.FastUp},
        {Key: "key_fast_down", DisplayKey: Strings["SettingsKeyBindingKeyFastDown"], Value: cfg.Keys.FastDown},
        {Key: "enter", DisplayKey: Strings["SettingsKeyBindingKeyEnter"], Value: cfg.Keys.Confirm},
        {Key: "back_cancel", DisplayKey: Strings["SettingsKeyBindingKeyBack"], Value: exitKeysStr},
        {Key: "add_comment", DisplayKey: Strings["SettingsKeyBindingComment"], Value: cfg.Keys.Comment},
        {Key: "search", DisplayKey: Strings["SettingsKeyBindingSearch"], Value: cfg.Keys.Search},
    }

	items := make([]list.Item, len(fields))
	for i := range fields {
		items[i] = fields[i]
	}

	l := list.New(items, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	l.Title = Strings["ConfigurationTitle"]
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return types.ConfigListEditor{
		Fields: fields,
		List:   l,
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
	for _, field := range m.configList.Fields {
        switch field.Key {
        case "key_up":
            m.cfg.Keys.Up = field.Value
        case "key_down":
            m.cfg.Keys.Down = field.Value
        case "key_fast_up":
            m.cfg.Keys.FastUp = field.Value
        case "key_fast_down":
            m.cfg.Keys.FastDown = field.Value
        case "enter":
            m.cfg.Keys.Confirm = field.Value
        case "back_cancel":
            m.cfg.Keys.Exit = strings.Split(field.Value, "/")
        case "add_comment":
            m.cfg.Keys.Comment = field.Value
        case "search":
            m.cfg.Keys.Search = field.Value

        // Add other config fields here if needed
        case "base_url":
            m.cfg.BaseURL = field.Value
        case "email":
            m.cfg.Email = field.Value
        case "api_token":
            m.cfg.APIToken = field.Value
        case "jql":
            m.cfg.JQL = field.Value
        case "language":
            m.cfg.Lang = field.Value
        }
    }
	if err := config.Save(m.cfg); err != nil {
		m.statusBar = "Error saving config: " + err.Error()
	}
}
