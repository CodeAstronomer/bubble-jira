package ui

import (
    "context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bubble-jira/config"
	"bubble-jira/jira"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
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

type menuItem struct {
	title   string
	enabled bool
}

const (
	MenuViewTasksTitle     = "View Jira Tasks"
	MenuSettingsTitle      = "Settings"
	MenuQuitTitle          = "Quit"
	MenuConfigTitle        = "Edit Config"
	MenuLicenceTitle       = "LICENSE"
	MenuBackTitle          = "Back to Main Menu"
)

func (m menuItem) Title() string       { return m.title }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return m.title }

// ---------- Loading / Fetching messages ----------

type issuesFetchedMsg struct {
	issues []jira.Issue
	err    error
}

type licenceLoadedMsg struct {
	content string
	err     error
}

type tickMsg time.Time

// ---------- Fetching model ----------

type fetchingModel struct {
	spinner   spinner.Model
	progress  progress.Model
	stages    []string
	currentStage int
	status    string
	error     string
	done      bool
	width     int
}

func newFetchingModel() fetchingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return fetchingModel{
		spinner: s,
		progress: p,
		stages: []string{
			"Connecting to Jira...",
			"Authenticating...",
			"Fetching tasks...",
			"Processing results...",
		},
		currentStage: 0,
		status: "Connecting to Jira...",
		width: 80,
	}
}

// ---------- Config Editor ----------

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton = focusedStyle.Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
	noStyle      = lipgloss.NewStyle()
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	checkMark    = successStyle.Render("✓")
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
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
	state            string
	menu             list.Model
	settings         list.Model
	tasksTable       table.Model
	configList       configListEditor
	configInput      configInputEditor
	fetching         fetchingModel
	statusBar        string
	editingFieldIdx  int
	configValidError string
	licenceContent   string
	licenceLoading   bool
}

// Lipgloss styles
var (
	menuStyle     = lipgloss.NewStyle().Padding(1, 2)
	tasksStyle    = lipgloss.NewStyle().Padding(1, 2)
	configStyle   = lipgloss.NewStyle().Padding(1, 2)
	fetchingStyle = lipgloss.NewStyle().Padding(2, 4)
	helpStyle     = blurredStyle
	tableBaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
)

func initialModel(cfg *config.Config, jc *jira.Client) model {
	menuItems := []list.Item{
		menuItem{title: MenuViewTasksTitle, enabled: cfg.IsValid()},
		menuItem{title: MenuSettingsTitle, enabled: true},
		menuItem{title: MenuQuitTitle, enabled: true},
	}
	menu := list.New(menuItems, list.NewDefaultDelegate(), 40, 15)
	menu.Title = "Main Menu"
	menu.SetShowHelp(true)
	menu.SetShowPagination(false)

	return model{
		cfg:     cfg,
		jc:      jc,
		state:   "menu",
		menu:    menu,
		fetching: newFetchingModel(),
	}
}

// ---------- Commands ----------

func fetchJiraTasksCmd(jc *jira.Client) tea.Cmd {
	return func() tea.Msg {
		issues, err := jc.FetchAssignedIssues(context.Background())
		return issuesFetchedMsg{issues: issues, err: err}
	}
}

func tickFetchCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadLicenceCmd() tea.Cmd {
	return func() tea.Msg {
		content, err := loadLicenceContent()
		return licenceLoadedMsg{content: content, err: err}
	}
}

func loadLicenceContent() (string, error) {
	// Try to load from local LICENSE file first
	execPath, err := os.Executable()
	if err == nil {
		projectRoot := filepath.Dir(execPath)
		localLicencePath := filepath.Join(projectRoot, "LICENSE")
		if data, err := ioutil.ReadFile(localLicencePath); err == nil {
			return string(data), nil
		}
	}

	// Try from current working directory
	if data, err := ioutil.ReadFile("LICENSE"); err == nil {
		return string(data), nil
	}

	// Try to fetch from GitHub
	githubURL := "https://raw.githubusercontent.com/DavidBachDerEchte/bubble-jira/main/LICENSE"
	resp, err := http.Get(githubURL)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			return string(data), nil
		}
	}

	return "License file not found. Please ensure LICENSE file exists in the project root or is available on GitHub.", nil
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
				menuItemSelected := selected.(menuItem)

				// Check if item is enabled
				if !menuItemSelected.enabled && menuItemSelected.title == MenuViewTasksTitle {
					m.configValidError = "⚠ Config incomplete! Please fill all fields in Settings > Edit Config first."
					return m, nil
				}

				switch menuItemSelected.title {
				case MenuViewTasksTitle:
					m.state = "fetching"
					m.fetching = newFetchingModel()
					m.configValidError = ""
					return m, tea.Batch(
						fetchJiraTasksCmd(m.jc),
						tickFetchCmd(),
						m.fetching.spinner.Tick,
					)
				case MenuSettingsTitle:
					m.state = "settings"
					m.configValidError = ""
					settingsItems := []list.Item{
						menuItem{title: MenuConfigTitle, enabled: true},
						menuItem{title: MenuLicenceTitle, enabled: true},
						menuItem{title: MenuBackTitle, enabled: true},
					}
					m.settings = list.New(settingsItems, list.NewDefaultDelegate(), 40, 15)
					m.settings.Title = "Settings"
					m.settings.SetShowHelp(true)
					m.settings.SetShowPagination(false)
				case MenuQuitTitle:
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case "settings":
		var cmd tea.Cmd
		m.settings, cmd = m.settings.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				selected := m.settings.SelectedItem()
				menuItemSelected := selected.(menuItem)

				switch menuItemSelected.title {
				case MenuConfigTitle:
					m.configList = newConfigListEditor(m.cfg)
					m.state = "config"
					return m, nil
				case MenuLicenceTitle:
					m.licenceLoading = true
					m.state = "licence"
					return m, loadLicenceCmd()
				case MenuBackTitle:
					m.state = "menu"
					return m, nil
				}
			} else if msg.String() == "q" || msg.String() == "esc" {
				m.state = "menu"
				return m, nil
			}
		}
		return m, cmd

	case "licence":
		switch msg := msg.(type) {
		case licenceLoadedMsg:
			m.licenceContent = msg.content
			m.licenceLoading = false
			return m, nil

		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
				m.state = "settings"
				m.licenceContent = ""
			}
		}
		return m, nil

	case "fetching":
		switch msg := msg.(type) {
		case tickMsg:
			// Advance progress stages
			if m.fetching.currentStage < len(m.fetching.stages)-1 {
				m.fetching.currentStage++
				percent := float64(m.fetching.currentStage) / float64(len(m.fetching.stages))
				progressCmd := m.fetching.progress.SetPercent(percent)
				m.fetching.status = m.fetching.stages[m.fetching.currentStage]
				return m, tea.Batch(tickFetchCmd(), progressCmd)
			}
			return m, nil

		case issuesFetchedMsg:
			if msg.err != nil {
				m.fetching.error = msg.err.Error()
				m.fetching.done = true
				m.fetching.progress.SetPercent(1.0)
				m.state = "fetching"
				return m, nil
			}

			// Create table rows from issues
			rows := make([]table.Row, len(msg.issues))
			for i, issue := range msg.issues {
				rows[i] = table.Row{
					issue.Key,
					truncateString(issue.Title, 40),
					issue.Status,
				}
			}

			columns := []table.Column{
				{Title: "Key", Width: 12},
				{Title: "Title", Width: 40},
				{Title: "Status", Width: 15},
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(15),
			)

			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)
			t.SetStyles(s)

			m.tasksTable = t
			m.fetching.progress.SetPercent(1.0)
			m.fetching.done = true
			m.fetching.status = "Complete!"
			m.state = "tasks"
			return m, nil

		case spinner.TickMsg:
			var cmd tea.Cmd
			m.fetching.spinner, cmd = m.fetching.spinner.Update(msg)
			return m, cmd

		case progress.FrameMsg:
			newProgress, cmd := m.fetching.progress.Update(msg)
			m.fetching.progress = newProgress.(progress.Model)
			return m, cmd

		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "esc" {
				m.state = "menu"
			}
		}
		return m, nil

	case "tasks":
		var cmd tea.Cmd
		m.tasksTable, cmd = m.tasksTable.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				m.state = "menu"
			case "enter":
				row := m.tasksTable.SelectedRow()
				if len(row) > 0 {
					return m, tea.Printf("Issue: %s\n", row[0])
				}
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
				m.state = "settings"
				// Refresh menu items enabled state
				settingsItems := []list.Item{
					menuItem{title: MenuConfigTitle, enabled: true},
					menuItem{title: MenuLicenceTitle, enabled: true},
					menuItem{title: MenuBackTitle, enabled: true},
				}
				m.settings = list.New(settingsItems, list.NewDefaultDelegate(), 40, 15)
				m.settings.Title = "Settings"
				m.settings.SetShowHelp(true)
				m.settings.SetShowPagination(false)
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
		view := m.menu.View()
		if m.configValidError != "" {
			view = view + "\n\n" + errorStyle.Render(m.configValidError)
		}
		return menuStyle.Render(view)
	case "settings":
		return menuStyle.Render(m.settings.View())
	case "licence":
		return menuStyle.Render(m.licenceView())
	case "tasks":
		return tasksStyle.Render(m.tasksTableView())
	case "config":
		return configStyle.Render(m.configListView())
	case "config-edit":
		return configStyle.Render(m.configInputView())
	case "fetching":
		return fetchingStyle.Render(m.fetchingView())
	default:
		return "Unknown state"
	}
}

func (m model) licenceView() string {
	var b strings.Builder

	if m.licenceLoading {
		b.WriteString("Loading licence information...\n")
		return b.String()
	}

	// Split content into lines and show with scrollable view
	lines := strings.Split(m.licenceContent, "\n")

	// Limit to 20 lines for terminal view, show first 20 lines
	maxLines := 1000
	if len(lines) > maxLines {
		for i := 0; i < maxLines; i++ {
			b.WriteString(lines[i])
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(warningStyle.Render(fmt.Sprintf("... (%d more lines)", len(lines)-maxLines)))
	} else {
		b.WriteString(m.licenceContent)
	}

	b.WriteString("\n\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press Q, ESC, or Enter to return to Settings"))

	return b.String()
}

func (m model) tasksTableView() string {
	var b strings.Builder
	b.WriteString(tableBaseStyle.Render(m.tasksTable.View()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Arrow keys: navigate | Q: back to menu | Enter: view issue"))
	return b.String()
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

func (m model) fetchingView() string {
	var b strings.Builder

	// Title
	b.WriteString("Fetching Jira Tasks\n")
	b.WriteString(strings.Repeat("─", 40))
	b.WriteString("\n\n")

	// Status with spinner
	spin := m.fetching.spinner.View()
	statusText := fmt.Sprintf("%s %s", spin, m.fetching.status)
	b.WriteString(statusText)
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(m.fetching.progress.View())
	b.WriteString("\n\n")

	// Error message if any
	if m.fetching.error != "" {
		b.WriteString(errorStyle.Render("✗ Error: " + m.fetching.error))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press Q or ESC to return to menu"))
	} else if m.fetching.done {
		b.WriteString(successStyle.Render(checkMark + " Tasks fetched successfully!"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press Q to return to menu or view tasks"))
	} else {
		b.WriteString(helpStyle.Render("Press Q or ESC to cancel"))
	}

	return b.String()
}

// ---------- Helper functions ----------

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}