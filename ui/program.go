package ui

import (
	"context"
	"encoding/json"
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
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

type contextMenuItem struct {
	title string
}

const (
	MenuViewTasksTitle     = "View Jira Tasks"
	MenuSettingsTitle      = "Settings"
	MenuQuitTitle          = "Quit"
	MenuConfigTitle        = "Edit Config"
	MenuLicenceTitle       = "LICENSE"
	MenuBackTitle          = "Back to Main Menu"
	ContextViewComments    = "View Comments"
)

func (m menuItem) Title() string       { return m.title }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return m.title }

func (cm contextMenuItem) Title() string       { return cm.title }
func (cm contextMenuItem) Description() string { return "" }
func (cm contextMenuItem) FilterValue() string { return cm.title }

// ---------- Loading / Fetching messages ----------

type issuesFetchedMsg struct {
	issues []jira.Issue
	err    error
}

type commentsFetchedMsg struct {
	comments []jira.Comment
	err      error
}

type licenceLoadedMsg struct {
	content string
	err     error
}

type tickMsg time.Time

// ---------- Comment body types ----------

type ContentNode struct {
	Type    string        `json:"type"`
	Content []ContentNode `json:"content,omitempty"`
	Text    string        `json:"text,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
}

type CommentBody struct {
	Type    string        `json:"type"`
	Version int           `json:"version"`
	Content []ContentNode `json:"content"`
}

// ---------- Fetching model ----------

type fetchingModel struct {
	spinner      spinner.Model
	progress     progress.Model
	stages       []string
	currentStage int
	status       string
	error        string
	done         bool
	width        int
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
		spinner:      s,
		progress:     p,
		stages:       []string{
			"Connecting to Jira...",
			"Authenticating...",
			"Fetching comments...",
			"Processing results...",
		},
		currentStage: 0,
		status:       "Connecting to Jira...",
		width:        80,
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
	authorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	timeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
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
	taskContextMenu  list.Model
	selectedIssue    *jira.Issue
	configList       configListEditor
	configInput      configInputEditor
	fetching         fetchingModel
	commentsViewport viewport.Model
	statusBar        string
	editingFieldIdx  int
	configValidError string
	licenceContent   string
	licenceLoading   bool
	comments         []jira.Comment
	commentsLoading  bool
	screenWidth      int
	screenHeight     int
}

// Lipgloss styles
var (
	menuStyle      = lipgloss.NewStyle().Padding(1, 2)
	tasksStyle     = lipgloss.NewStyle().Padding(1, 2)
	configStyle    = lipgloss.NewStyle().Padding(1, 2)
	fetchingStyle  = lipgloss.NewStyle().Padding(2, 4)
	helpStyle      = blurredStyle
	tableBaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	commentsStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0)
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

	vp := viewport.New(200, 50)
	vp.Style = commentsStyle

	return model{
		cfg:              cfg,
		jc:               jc,
		state:            "menu",
		menu:             menu,
		fetching:         newFetchingModel(),
		commentsViewport: vp,
		screenWidth:      200,
		screenHeight:     50,
	}
}

// ---------- Commands ----------

func fetchJiraTasksCmd(jc *jira.Client) tea.Cmd {
	return func() tea.Msg {
		issues, err := jc.FetchAssignedIssues(context.Background())
		return issuesFetchedMsg{issues: issues, err: err}
	}
}

func fetchCommentsCmd(jc *jira.Client, issueKey string) tea.Cmd {
	return func() tea.Msg {
		comments, err := jc.FetchComments(context.Background(), issueKey)
		return commentsFetchedMsg{comments: comments, err: err}
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

// ---------- Helper Functions ----------

func parseContentNodes(nodes []ContentNode, result *strings.Builder, depth int) {
	for _, node := range nodes {
		switch node.Type {
		case "paragraph":
			for _, child := range node.Content {
				switch child.Type {
				case "text":
					result.WriteString(child.Text)
				case "mention":
					if text, ok := child.Attrs["text"].(string); ok {
						result.WriteString(text)
					}
				case "hardBreak":
					result.WriteString("\n")
				}
			}
			result.WriteString("\n")

		case "bulletList", "orderedList":
			for _, child := range node.Content {
				if child.Type == "listItem" {
					result.WriteString("• ")
					parseContentNodes(child.Content, result, depth+1)
				}
			}

		case "codeBlock":
			result.WriteString("```\n")
			for _, child := range node.Content {
				if child.Type == "text" {
					result.WriteString(child.Text)
				}
			}
			result.WriteString("\n```\n")

		case "heading":
			if level, ok := node.Attrs["level"].(float64); ok {
				for i := 0; i < int(level); i++ {
					result.WriteString("#")
				}
				result.WriteString(" ")
			}
			for _, child := range node.Content {
				if child.Type == "text" {
					result.WriteString(child.Text)
				}
			}
			result.WriteString("\n")

		case "text":
			result.WriteString(node.Text)

		default:
			// Recursively parse unknown types
			if len(node.Content) > 0 {
				parseContentNodes(node.Content, result, depth)
			}
		}
	}
}

func parseCommentBody(bodyJSON string) string {
	var commentBody CommentBody
	if err := json.Unmarshal([]byte(bodyJSON), &commentBody); err != nil {
		// If it's not valid JSON, return as-is
		return bodyJSON
	}

	var result strings.Builder

	parseContentNodes(commentBody.Content, &result, 0)

	return strings.TrimSpace(result.String())
}

func formatTime(timeStr string) string {
	// Parse the time string from Jira (e.g., "2025-10-28T09:31:21.319+0100")
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", timeStr)
	if err != nil {
		// Try alternative format
		t, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return timeStr
		}
	}
	return t.Format("2006-01-02 15:04 MST")
}

func renderCommentsToMarkdown(comments []jira.Comment) string {
	var result strings.Builder
	result.WriteString("# Issue Comments\n\n")

	if len(comments) == 0 {
		result.WriteString("No comments found.\n")
		return result.String()
	}

	for i, comment := range comments {
		// Main comment header
		result.WriteString(fmt.Sprintf("## Comment #%d\n\n", i+1))

		// Author and timestamp
		result.WriteString(fmt.Sprintf("**👤 Author:** %s\n\n", comment.Author.DisplayName))
		result.WriteString(fmt.Sprintf("**📧 Email:** %s\n\n", comment.Author.EmailAddress))
		result.WriteString(fmt.Sprintf("**⏰ Posted:** %s\n\n", formatTime(comment.Created)))

		// If updated is different from created, show update time
		if comment.Updated != comment.Created {
			result.WriteString(fmt.Sprintf("**📝 Last Updated:** %s\n\n", formatTime(comment.Updated)))
		}

		result.WriteString("---\n\n")

		// Comment body
		body := parseCommentBody(comment.BodyJSON)
		result.WriteString(body)
		result.WriteString("\n\n")

		// Process replies if they exist
		if len(comment.Replies) > 0 {
			result.WriteString("### 💬 Replies:\n\n")
			for j, reply := range comment.Replies {
				result.WriteString(fmt.Sprintf("#### Reply #%d\n\n", j+1))
				result.WriteString(fmt.Sprintf("**👤 Author:** %s\n\n", reply.Author.DisplayName))
				result.WriteString(fmt.Sprintf("**📧 Email:** %s\n\n", reply.Author.EmailAddress))
				result.WriteString(fmt.Sprintf("**⏰ Posted:** %s\n\n", formatTime(reply.Created)))

				if reply.Updated != reply.Created {
					result.WriteString(fmt.Sprintf("**📝 Last Updated:** %s\n\n", formatTime(reply.Updated)))
				}

				result.WriteString("---\n\n")
				replyBody := parseCommentBody(reply.BodyJSON)
				result.WriteString("> " + strings.ReplaceAll(replyBody, "\n", "\n> "))
				result.WriteString("\n\n")
			}
		}

		result.WriteString("\n")
		result.WriteString(strings.Repeat("═", 95))
		result.WriteString("\n\n")
	}

	return result.String()
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
					// Create context menu with only View Comments option
					contextItems := []list.Item{
						contextMenuItem{title: ContextViewComments},
					}
					m.taskContextMenu = list.New(contextItems, list.NewDefaultDelegate(), 30, 10)
					m.taskContextMenu.Title = "Actions"
					m.taskContextMenu.SetShowHelp(true)
					m.taskContextMenu.SetShowPagination(false)

					key := row[0]
					m.selectedIssue = &jira.Issue{
						Key:    key,
						Title:  row[1],
						Status: row[2],
					}

					m.state = "task-context"
					return m, nil
				}
			}
		}
		return m, cmd

	case "task-context":
		var cmd tea.Cmd
		m.taskContextMenu, cmd = m.taskContextMenu.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				selected := m.taskContextMenu.SelectedItem()
				menuItem := selected.(contextMenuItem)

				if menuItem.title == ContextViewComments && m.selectedIssue != nil {
					m.commentsLoading = true
					m.fetching = newFetchingModel()
					m.state = "comments-fetching"
					return m, tea.Batch(
						fetchCommentsCmd(m.jc, m.selectedIssue.Key),
						tickFetchCmd(),
						m.fetching.spinner.Tick,
					)
				}
				m.state = "tasks"
				m.selectedIssue = nil
				return m, nil

			case "q", "esc":
				m.state = "tasks"
				m.selectedIssue = nil
				return m, nil
			}
		}
		return m, cmd

	case "comments-fetching":
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

		case commentsFetchedMsg:
			if msg.err != nil {
				m.fetching.error = msg.err.Error()
				m.fetching.done = true
				m.fetching.progress.SetPercent(1.0)
				m.state = "comments-fetching"
				return m, nil
			}

			m.comments = msg.comments
			m.commentsLoading = false

			// Render comments to markdown
			markdownContent := renderCommentsToMarkdown(m.comments)

			// Render markdown using glamour
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.screenWidth-4),
			)
			if err != nil {
				m.fetching.error = err.Error()
				m.state = "comments-fetching"
				return m, nil
			}

			renderedContent, err := renderer.Render(markdownContent)
			if err != nil {
				m.fetching.error = err.Error()
				m.state = "comments-fetching"
				return m, nil
			}

			m.commentsViewport.SetContent(renderedContent)
			m.state = "comments-view"
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
				m.state = "tasks"
			}
		}
		return m, nil

	case "comments-view":
		var cmd tea.Cmd
		m.commentsViewport, cmd = m.commentsViewport.Update(msg)

		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.screenWidth = msg.Width
			m.screenHeight = msg.Height
			m.commentsViewport.Width = msg.Width
			m.commentsViewport.Height = msg.Height - 2

		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.state = "tasks"
				m.selectedIssue = nil
				return m, nil
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
	case "task-context":
		return tasksStyle.Render(m.taskContextView())
	case "comments-fetching":
		return fetchingStyle.Render(m.fetchingView())
	case "comments-view":
		return m.commentsView()
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
	b.WriteString(helpStyle.Render("Arrow keys: navigate | Q: back to menu | Enter: open actions"))
	return b.String()
}

func (m model) taskContextView() string {
	var b strings.Builder
	if m.selectedIssue != nil {
		b.WriteString(fmt.Sprintf("Selected: %s - %s\n\n", m.selectedIssue.Key, m.selectedIssue.Title))
	}
	b.WriteString(m.taskContextMenu.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter: select | Q/ESC: back"))
	return b.String()
}

func (m model) commentsView() string {
	var b strings.Builder
	b.WriteString(m.commentsViewport.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: Scroll | Q/ESC: Back to tasks"))
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
	b.WriteString("Fetching Jira Data\n")
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
		b.WriteString(helpStyle.Render("Press Q or ESC to return"))
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