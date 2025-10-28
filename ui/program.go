package ui

import (
	"fmt"
	"strings"

	"bubble-jira/config"
	"bubble-jira/jira"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Program wraps the bubbletea program
type Program struct {
	p *tea.Program
}

func NewProgram(cfg *config.Config, jc *jira.Client) *Program {
	m := initialModel(cfg, jc)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return &Program{p: p}
}

func (pr *Program) Start() error {
	return pr.p.Start()
}

func (pr *Program) Quit() {
	pr.p.Send(tea.Quit)
}

// ---------- Menu & Models ----------

type menuItem string

const (
	MenuViewTasks  menuItem = "View Jira Tasks"
	MenuEditConfig menuItem = "Edit Config"
	MenuQuit       menuItem = "Quit"
)

func (m menuItem) Title() string       { return string(m) }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return string(m) }

// ---------- Config Editor ----------

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton = focusedStyle.Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
	noStyle      = lipgloss.NewStyle()
)

type configField struct {
	key   string
	value string
}

func (cf configField) Title() string       { return cf.key }
func (cf configField) Description() string { return cf.value }
func (cf configField) FilterValue() string { return cf.key }

type configListEditor struct {
	fields []configField
	list   list.Model
}

func newConfigListEditor(cfg *config.Config) configListEditor {
	fields := []configField{
		{key: "Base URL", value: cfg.BaseURL},
		{key: "Email", value: cfg.Email},
		{key: "API Token", value: cfg.APIToken},
		{key: "JQL Query", value: cfg.JQL},
	}

	items := make([]list.Item, len(fields))
	for i, f := range fields {
		items[i] = f
	}

	l := list.New(items, list.NewDefaultDelegate(), 50, 20)
	l.Title = "Configuration"
	l.SetShowHelp(true)
	l.SetShowPagination(false)

	return configListEditor{
		fields: fields,
		list:   l,
	}
}

type configInputEditor struct {
	input     textinput.Model
	key       string
	focusSave bool
}

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

// ---------- Main Model ----------

type model struct {
	cfg              *config.Config
	jc               *jira.Client
	state            string // "menu" | "tasks" | "config" | "config-edit"
	menu             list.Model
	tasks            list.Model
	configList       configListEditor
	configInput      configInputEditor
	statusBar        string
	editingFieldIdx  int
}

// Lipgloss styles
var (
	menuStyle     = lipgloss.NewStyle().Padding(1, 2)
	tasksStyle    = lipgloss.NewStyle().Padding(1, 2)
	configStyle   = lipgloss.NewStyle().Padding(1, 2)
	helpStyle     = blurredStyle
)

func initialModel(cfg *config.Config, jc *jira.Client) model {
	menuItems := []list.Item{
		MenuViewTasks,
		MenuEditConfig,
		MenuQuit,
	}
	menu := list.New(menuItems, list.NewDefaultDelegate(), 40, 15)
	menu.Title = "Main Menu"
	menu.SetShowHelp(true)
	menu.SetShowPagination(false)

	return model{
		cfg:   cfg,
		jc:    jc,
		state: "menu",
		menu:  menu,
	}
}

// ---------- Bubble Tea implementation ----------

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case "menu":
		var cmd tea.Cmd
		m.menu, cmd = m.menu.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				selected := m.menu.SelectedItem()
				switch selected.(menuItem) {
				case MenuViewTasks:
					issues, err := m.jc.FetchAssignedIssues(nil)
					if err != nil {
						m.statusBar = "Error fetching tasks: " + err.Error()
						return m, nil
					}
					items := make([]list.Item, len(issues))
					for i, is := range issues {
						items[i] = item{Issue: is}
					}
					tasksList := list.New(items, list.NewDefaultDelegate(), 0, 0)
					tasksList.Title = "Assigned Jira Tasks"
					tasksList.SetShowHelp(false)
					tasksList.SetShowPagination(false)
					m.tasks = tasksList
					m.state = "tasks"
				case MenuEditConfig:
					m.configList = newConfigListEditor(m.cfg)
					m.state = "config"
				case MenuQuit:
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case "tasks":
		var cmd tea.Cmd
		m.tasks, cmd = m.tasks.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" {
				m.state = "menu"
			}
		}
		return m, cmd

	case "config":
		var cmd tea.Cmd
		m.configList.list, cmd = m.configList.list.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.editingFieldIdx = m.configList.list.Index()
				field := m.configList.fields[m.editingFieldIdx]
				m.configInput = newConfigInputEditor(field.key, field.value)
				m.state = "config-edit"
				return m, nil

			case "q", "esc":
				m.state = "menu"
				return m, nil
			}
		}
		return m, cmd

	case "config-edit":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab", "down":
				m.configInput.focusSave = !m.configInput.focusSave
				if m.configInput.focusSave {
					m.configInput.input.Blur()
					m.configInput.input.PromptStyle = noStyle
					m.configInput.input.TextStyle = noStyle
				} else {
					m.configInput.input.Focus()
					m.configInput.input.PromptStyle = focusedStyle
					m.configInput.input.TextStyle = focusedStyle
				}
				return m, nil

			case "up":
				if m.configInput.focusSave {
					m.configInput.focusSave = false
					m.configInput.input.Focus()
					m.configInput.input.PromptStyle = focusedStyle
					m.configInput.input.TextStyle = focusedStyle
				}
				return m, nil

			case "enter":
				if m.configInput.focusSave {
					// Save and return to config list
					m.configList.fields[m.editingFieldIdx].value = m.configInput.input.Value()
					m.updateConfig()
					m.configList = newConfigListEditor(m.cfg)
					m.state = "config"
					return m, nil
				}

			case "esc":
				// Cancel and return to config list without saving
				m.configList = newConfigListEditor(m.cfg)
				m.state = "config"
				return m, nil
			}

			if !m.configInput.focusSave {
				var cmd tea.Cmd
				m.configInput.input, cmd = m.configInput.input.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

func (m *model) updateConfig() {
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

func (m model) View() string {
	switch m.state {
	case "menu":
		return menuStyle.Render(m.menu.View())
	case "tasks":
		return tasksStyle.Render(m.tasks.View())
	case "config":
		return configStyle.Render(m.configListView())
	case "config-edit":
		return configStyle.Render(m.configInputView())
	default:
		return "Unknown state"
	}
}

func (m model) configListView() string {
	var b strings.Builder
	b.WriteString(m.configList.list.View())
	return b.String()
}

func (m model) configInputView() string {
	var b strings.Builder

	b.WriteString("Editing: ")
	b.WriteString(focusedStyle.Render(m.configInput.key))
	b.WriteString("\n\n")

	b.WriteString("Value:\n")
	b.WriteString(m.configInput.input.View())
	b.WriteString("\n\n")

	button := &blurredButton
	if m.configInput.focusSave {
		button = &focusedButton
	}
	b.WriteString(fmt.Sprintf("%s\n\n", *button))

	b.WriteString(helpStyle.Render("Arrow Up/Down: navigate | Enter: confirm | Esc: cancel"))

	return b.String()
}

// ---------- List item for Jira tasks ----------

type item struct {
	Issue jira.Issue
}

func (i item) Title() string       { return fmt.Sprintf("[%s] %s", i.Issue.Key, i.Issue.Title) }
func (i item) Description() string { return i.Issue.Status }
func (i item) FilterValue() string { return i.Issue.Key + " " + i.Issue.Title }