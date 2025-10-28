package ui

import (
	"bubble-jira/jira"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// updateMenu handles menu state updates
func (m model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			selected := m.menu.SelectedItem()
			menuItemSelected := selected.(menuItem)

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
}

// updateSettings handles settings state updates
func (m model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateLicence handles license state updates
func (m model) updateLicence(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateFetching handles fetching state updates
func (m model) updateFetching(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
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
}

// updateTasks handles tasks state updates
func (m model) updateTasks(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateTaskContext handles task context menu updates
func (m model) updateTaskContext(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateCommentsFetching handles comments fetching state updates
func (m model) updateCommentsFetching(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
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

		markdownContent := renderCommentsToMarkdown(m.comments)

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
}

// updateCommentsView handles comments view updates
func (m model) updateCommentsView(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateConfigList handles config list state updates
func (m model) updateConfigList(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

// updateConfigEdit handles config edit state updates
func (m model) updateConfigEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.configList.fields[m.editingFieldIdx].value = m.configInput.input.Value()
				m.updateConfigFields()
				m.configList = newConfigListEditor(m.cfg)
				m.state = "config"
				return m, nil
			}

		case "esc":
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
	return m, nil
}