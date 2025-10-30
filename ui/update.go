package ui

import (
	"bubble-jira/jira"
	"bubble-jira/config"
    "fmt"
    "time"
    "regexp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/glamour"
    "github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

var languageRegex = regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`)

func contains(slice []string, val string) bool {
    for _, s := range slice {
        if s == val {
            return true
        }
    }
    return false
}

// updateMenu handles menu state updates
func (m model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == keyEnter {
			selected := m.menu.SelectedItem()
			menuItemSelected := selected.(menuItem)

			if !menuItemSelected.enabled && menuItemSelected.title == Strings["MenuViewTasksTitle"] {
				m.configValidError = "⚠ Config incomplete! Please fill all fields in Settings > Edit Config first."
				return m, nil
			}

			switch menuItemSelected.title {
			case Strings["MenuViewTasksTitle"]:
				m.state = "fetching"
				m.fetching = newFetchingModel()
				m.configValidError = ""
				return m, tea.Batch(
					fetchJiraTasksCmd(m.jc),
					tickFetchCmd(),
					m.fetching.spinner.Tick,
				)
			case Strings["MenuSettingsTitle"]:
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
			case Strings["MenuQuitTitle"]:
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
		if msg.String() == keyEnter {
			selected := m.settings.SelectedItem()
			menuItemSelected := selected.(menuItem)

			switch menuItemSelected.title {
			case Strings["MenuConfigTitle"]:
				m.configList = newConfigListEditor(m.cfg)
				m.editingMode = "config"
				m.state = "config"
				return m, nil
			case Strings["MenuLicenceTitle"]:
				m.licenceLoading = true
				m.state = "licence"
				return m, loadLicenceCmd()
			case Strings["MenuKeybindingTitle"]:
                m.configList = newKeybindingEditor(m.cfg)
				m.editingMode = "keybindings"
                m.state = "config"
                return m, nil
			case Strings["MenuBackTitle"]:
				m.state = "menu"
				return m, nil
			}
		} else if contains(exitKeys, msg.String()) {
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
        m.licenceOffset = 0
        return m, nil

    case tea.KeyMsg:
        switch {
        case contains(exitKeys, msg.String()):
            m.state = "settings"
            m.licenceContent = ""
        case msg.String() == keyUp:
            m.licenceOffset--
        case msg.String() == keyDown:
            m.licenceOffset++
        case msg.String() == keyFastUp:
            m.licenceOffset -= terminalHeight / 2
        case msg.String() == keyFastDown:
            m.licenceOffset += terminalHeight / 2
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

        	// Calculate column widths dynamically
        	keyWidth := int(float64(terminalWidth) * 0.12)
        	if keyWidth < 12 {
        		keyWidth = 12
        	}

        	statusWidth := int(float64(terminalWidth) * 0.15)
        	if statusWidth < 10 {
        		statusWidth = 10
        	}

        	titleWidth := terminalWidth - keyWidth - statusWidth
        	if titleWidth < 20 { // minimum title width
        		titleWidth = 20
        	}

        	// Populate table rows with truncated title based on column width
        	for i, issue := range msg.issues {
        		rows[i] = table.Row{
        			issue.Key,
        			truncateString(issue.Title, titleWidth/2), // rough char approximation
        			issue.Status,
        		}
        	}

        	columns := []table.Column{
        		{Title: Strings["TaskListItemKey"], Width: keyWidth},
        		{Title: Strings["TaskListItemTitle"], Width: titleWidth},
        		{Title: Strings["TaskListItemStatus"], Width: statusWidth},
        	}

        	t := table.New(
        		table.WithColumns(columns),
        		table.WithRows(rows),
        		table.WithFocused(true),
        		table.WithHeight(taskViewHeight),
        	)

        	t.SetStyles(table.Styles{
        		Header:   headerStyle,
        		Selected: selectedStyle,
        	})

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
		if contains(exitKeys, msg.String()) {
			m.state = "menu"
		}
	}
	return m, nil
}

// updateTasks handles tasks state updates
func (m model) updateTasks(msg tea.Msg) (tea.Model, tea.Cmd) {
    oldCursor := m.tasksTable.Cursor()

    switch msg := msg.(type) {
    case hoverTimeoutMsg:
        if msg.taskKey == m.hoveredTaskKey && !m.fetchingComments {
            if _, exists := m.cachedComments[msg.taskKey]; !exists {
                m.fetchingComments = true
                return m, fetchCommentsBackgroundCmd(m.jc, msg.taskKey)
            }
        }
        return m, nil

    case commentsCachedMsg:
        m.fetchingComments = false
        if msg.err == nil {
            m.cachedComments[msg.taskKey] = msg.comments
        }
        return m, nil

    case tea.KeyMsg:
        switch {
        case contains(exitKeys, msg.String()):
            if m.hoverTimer != nil {
                m.hoverTimer.Stop()
                m.hoverTimer = nil
            }
            m.hoveredTaskKey = ""
            m.state = "menu"
            return m, nil

        case msg.String() == " ":
            row := m.tasksTable.SelectedRow()
            if len(row) > 0 {
                key := row[0]
                m.selectedIssue = &jira.Issue{
                    Key:    key,
                    Title:  row[1],
                    Status: row[2],
                }

                if m.hoverTimer != nil {
                    m.hoverTimer.Stop()
                    m.hoverTimer = nil
                }

                if cachedComments, exists := m.cachedComments[key]; exists {
                    m.comments = cachedComments
                    m.commentsLoading = false

                    markdownContent := renderCommentsToMarkdown(m.comments)
                    renderer, err := glamour.NewTermRenderer(
                        glamour.WithAutoStyle(),
                        glamour.WithWordWrap(m.screenWidth-4),
                    )
                    if err != nil {
                        m.state = "tasks"
                        return m, nil
                    }

                    renderedContent, err := renderer.Render(markdownContent)
                    if err != nil {
                        m.state = "tasks"
                        return m, nil
                    }

                    m.commentsViewport.SetContent(renderedContent)
                    m.state = "comments-view"
                    return m, nil
                } else if m.fetchingComments && m.hoveredTaskKey == key {
                    m.commentsLoading = true
                    m.fetching = newFetchingModel()
                    m.state = "comments-fetching"
                    return m, tea.Batch(
                        tickFetchCmd(),
                        m.fetching.spinner.Tick,
                    )
                } else {
                    m.commentsLoading = true
                    m.fetching = newFetchingModel()
                    m.state = "comments-fetching"
                    return m, tea.Batch(
                        fetchCommentsCmd(m.jc, key),
                        tickFetchCmd(),
                        m.fetching.spinner.Tick,
                    )
                }
            }
            return m, nil

        case msg.String() == keyEnter:
        	row := m.tasksTable.SelectedRow()
        	if len(row) == 0 {
        		return m, nil
        	}

        	key := row[0]

        	// Initialize jiraStatusInput with only the key
            m.jiraStatusInput = struct {
                key      string
                input    textinput.Model
                focusSave bool
                cursor   int
                choice   string
            }{
                key: key,
            }

        	// Create task settings menu
        	taskSettingsItems := []list.Item{
            	menuItem{title: Strings["SetState"], enabled: true},
            	menuItem{title: Strings["CopyTitle"], enabled: true},
            	menuItem{title: Strings["MenuBackTitle"], enabled: true},
            }

            m.taskSettings = list.New(taskSettingsItems, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
            m.taskSettings.Title = fmt.Sprintf(Strings["TaskTitle"] + ": %s", key)
            m.taskSettings.SetShowHelp(true)
            m.taskSettings.SetShowPagination(false)
            m.state = "task-settings-list"

        	return m, nil
        }
    }

    var cmd tea.Cmd
    m.tasksTable, cmd = m.tasksTable.Update(msg)
    newCursor := m.tasksTable.Cursor()

    if oldCursor != newCursor {
        if m.hoverTimer != nil {
            m.hoverTimer.Stop()
        }

        // Neuen Timer starten
        row := m.tasksTable.SelectedRow()
        if len(row) > 0 {
            taskKey := row[0]
            m.hoveredTaskKey = taskKey
            return m, hoverTimeoutCmd(taskKey, time.Duration(autoFetchTimeSec)*time.Second)
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
		switch {
		case msg.String() == keyEnter:
			selected := m.taskContextMenu.SelectedItem()
			menuItem := selected.(contextMenuItem)

			if menuItem.title == Strings["ContextViewComments"] && m.selectedIssue != nil {
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

		case contains(exitKeys, msg.String()):
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
		// Increment progress stage if not finished
		if m.fetching.currentStage < len(m.fetching.stages)-1 {
			m.fetching.currentStage++
			m.fetching.status = m.fetching.stages[m.fetching.currentStage]
			percent := float64(m.fetching.currentStage+1) / float64(len(m.fetching.stages))
			progressCmd := m.fetching.progress.SetPercent(percent)
			return m, tea.Batch(tickFetchCmd(), progressCmd)
		}

        // Progress is done
        if m.fetchCommentsAfterProgress {
            m.fetchCommentsAfterProgress = false
            return m, fetchCommentsCmd(m.jc, m.addCommentInput.issueKey)
        }

		// Only quit automatically if this animation came from statusCodeMsg
		if m.fetching.fromStatus {
			fmt.Printf("\nThis window will close in %d seconds.\n", closeAfterSec)
			m.fetching.fromStatus = false // reset flag
			return m, tea.Tick(time.Duration(closeAfterSec)*time.Second, func(time.Time) tea.Msg {
				return quitAfterDelayMsg{}
			})
		}
		return m, nil

	case statusCodeMsg:
		if msg.Code >= 200 && msg.Code < 300 {
			m.fetching.currentStage = 0
			m.fetching.progress.SetPercent(0.0)
			m.fetching.status = m.fetching.stages[0]
			m.fetching.fromStatus = true

			// Set flag to fetch comments after progress finishes
            if m.state == "comments-fetching" {
                m.fetchCommentsAfterProgress = true
            }

			return m, tickFetchCmd()
		}

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

	case quitAfterDelayMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if contains(exitKeys, msg.String()) {
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
		switch {

		case msg.String() == "n": //keyNewComment
            row := m.tasksTable.SelectedRow()
            if len(row) == 0 {
                return m, nil
            }
            key := row[0]

            ti := textinput.New()
            ti.Placeholder = "Write your comment..."
            ti.Focus()
            ti.CharLimit = 1000
            ti.Width = terminalWidth - 10

            m.addCommentInput = struct {
                input     textinput.Model
                focusSend bool
                issueKey  string
            }{
                input:     ti,
                focusSend: false,
                issueKey:  key,
            }

            m.state = "add-comment"
            return m, nil

		case contains(exitKeys, msg.String()):
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
		switch {
		case msg.String() == keyEnter:
			m.editingFieldIdx = m.configList.list.Index()
			field := m.configList.fields[m.editingFieldIdx]
			m.configInput = newConfigInputEditor(field.key, field.value)
			m.state = "config-edit"
			return m, nil

		case contains(exitKeys, msg.String()):
            m.state = "settings"
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
            return m, nil
		}
	}
	return m, cmd
}

// updateConfigEdit handles config edit state updates
func (m model) updateConfigEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == keyDown:
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

		case msg.String() == keyUp:
			if m.configInput.focusSave {
				m.configInput.focusSave = false
				m.configInput.input.Focus()
				m.configInput.input.PromptStyle = focusedStyle
				m.configInput.input.TextStyle = focusedStyle
			}
			return m, nil


		case msg.String() == keyEnter:
            if m.configInput.focusSave {
                value := m.configInput.input.Value()

                if m.editingMode == "keybindings" && m.configInput.key == Strings["language"] {
                    if !languageRegex.MatchString(value) {
                        m.configInput.focusSave = true
                        return m, nil
                    }

                    isAllowed := false
                    for _, lang := range allowedLanguages {
                        if value == lang {
                            isAllowed = true
                            break
                        }
                    }

                    if !isAllowed {
                        m.configInput.focusSave = true
                        return m, nil
                    }

                    m.cfg.Lang = value
                    if err := config.Save(m.cfg); err != nil {
                        m.configInput.focusSave = true
                        return m, nil
                    }

                    Lang = value
                    loadStrings(Lang)

                    m.configValidError = ""
                }

                m.configList.fields[m.editingFieldIdx].value = value
                m.updateConfigFields()

                if m.editingMode == "keybindings" {
                    m.configList = newKeybindingEditor(m.cfg)
                } else {
                    m.configList = newConfigListEditor(m.cfg)
                }
                m.state = "config"
                return m, nil
            }

		case contains(exitKeys, msg.String()):
			if m.editingMode == "keybindings" {
				m.configList = newKeybindingEditor(m.cfg)
			} else {
				m.configList = newConfigListEditor(m.cfg)
			}
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

// updateSettings handles settings state updates
func (m model) updateTaskSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.taskSettings, cmd = m.taskSettings.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == keyEnter {
			selected := m.taskSettings.SelectedItem()
			if selected == nil {
				return m, nil
			}

			menuItemSelected := selected.(menuItem)

			switch menuItemSelected.title {
			case Strings["SetState"]:
				m.state = "task-status"
				return m, nil

			case Strings["CopyTitle"]:
                if m.jiraStatusInput.key != "" {
                    // Copy the task key to clipboard
                    err := clipboard.WriteAll("["+m.jiraStatusInput.key+"]")
                    if err == nil {
                        m.statusMessage = "Task key copied to clipboard!"
                        fmt.Printf("Task key copied to clipboard!")
                    } else {
                        m.statusMessage = "Failed to copy task key!"
                        fmt.Printf("Task key copied to clipboard!")
                    }
                }
                return m, nil

			case Strings["MenuBackTitle"]:
				m.state = "tasks"
				return m, nil
			}
		} else if contains(exitKeys, msg.String()) {
			m.state = "tasks"
			return m, nil
		}
	}

	return m, cmd
}

// updateTaskStatus
func (m model) updateTaskStatus(msg tea.Msg) (tea.Model, tea.Cmd) {
	statuses := []string{
    	Strings["Open"],
    	Strings["CurrentlyInProgress"],
    	Strings["Done"],
    	Strings["Reopened"],
    	Strings["Closed"],
    	Strings["Backlog"],
    	Strings["QM"],
    	Strings["Waiting"],
    	Strings["Staging"],
    }

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case keyDown:
			if !m.jiraStatusInput.focusSave {
				m.jiraStatusInput.cursor++
				if m.jiraStatusInput.cursor >= len(statuses) {
					m.jiraStatusInput.cursor = 0
				}
			} else {
				m.jiraStatusInput.focusSave = false
			}
			return m, nil

		case keyUp:
			if !m.jiraStatusInput.focusSave {
				m.jiraStatusInput.cursor--
				if m.jiraStatusInput.cursor < 0 {
					m.jiraStatusInput.cursor = len(statuses) - 1
				}
			} else {
				m.jiraStatusInput.focusSave = false
			}
			return m, nil

		case keyEnter:
			selectedStatus := statuses[m.jiraStatusInput.cursor]
			m.jiraStatusInput.choice = selectedStatus
			m.state = "comments-fetching"
            var selectedID = JiraStatusMap[selectedStatus]
			return m, fetchStatusCmd(m.jc, m.jiraStatusInput.key, selectedID)

		default:
			if contains(exitKeys, msg.String()) {
				m.state = "task-settings-list"
				return m, nil
			}
		}

	case quitAfterDelayMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) updateAddCommentInput(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case keyEnter:
            if m.addCommentInput.focusSend {
                text := m.addCommentInput.input.Value()
                if len(text) == 0 || len(text) > 1000 {
                    return m, nil // do not submit invalid comment
                }

                m.fetching = newFetchingModel()
                m.state = "comments-fetching"
                return m, createNewComment(m.jc, m.addCommentInput.issueKey, text)

            } else {
                // toggle focus to send button
                m.addCommentInput.focusSend = true
                m.addCommentInput.input.Blur()
            }
            return m, nil

        case keyUp, keyDown:
            m.addCommentInput.focusSend = !m.addCommentInput.focusSend
            if !m.addCommentInput.focusSend {
                m.addCommentInput.input.Focus()
            } else {
                m.addCommentInput.input.Blur()
            }
            return m, nil

        case "esc":
            m.state = "tasks"
            return m, nil
        }

    case tea.WindowSizeMsg:
        m.screenWidth = msg.Width
        m.screenHeight = msg.Height
    }

    // Update text input if focused
    if !m.addCommentInput.focusSend {
        var cmd tea.Cmd
        m.addCommentInput.input, cmd = m.addCommentInput.input.Update(msg)
        return m, cmd
    }

    return m, nil
}