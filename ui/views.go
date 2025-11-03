package ui

import (
    "strings"
	"github.com/charmbracelet/lipgloss"
)

// licenceView renders the licence view
func (m model) licenceView() string {
    if m.licenceLoading {
        return Strings["LoadLicense"]
    }

    lines := strings.Split(m.licenceContent, "\n")

    // Calculate visible lines
    visibleLines := terminalHeight - 2 // leave 2 lines for instructions
    if m.licenceOffset > len(lines)-visibleLines {
        m.licenceOffset = max(0, len(lines)-visibleLines)
    }
    if m.licenceOffset < 0 {
        m.licenceOffset = 0
    }

    end := m.licenceOffset + visibleLines
    if end > len(lines) {
        end = len(lines)
    }
    visible := lines[m.licenceOffset:end]

    content := strings.Join(visible, "\n")
    footer := lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Scroll"]+" • "+ keyMap[keyFastUp]+"/"+ keyMap[keyFastDown] +": "+Strings["FastScroll"]+" • "+ keyExitKeysStr +": "+Strings["Back"])

    // Render with terminal width and height, but no forced wrapping
    return lipgloss.NewStyle().
        Width(terminalWidth).
        Height(terminalHeight).
        Render(content + "\n\n" + footer)
}

// tasksTableView renders the tasks table view
func (m model) tasksTableView() string {
	var content strings.Builder
	content.Grow(512)
	content.WriteString(tableBaseStyle.Render(m.tasksTable.View()))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Navigate"]+" • ␣: "+Strings["Comments"]+" • " + keyMap[keyEnter] +": "+Strings["Select"]+" • "+keySearch+": "+Strings["Search"]+" • "+ keyExitKeysStr +": "+Strings["Back"]))
	return content.String()
}

// taskContextView renders the task context menu view
func (m model) taskContextView() string {
	return m.taskContextMenu.View()
}

// fetchingView renders the fetching view
func (m model) fetchingView() string {
	if m.fetching.error != "" {
		return lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding).Render(
			"Error: " + m.fetching.error,
		)
	}

	var content strings.Builder
    content.Grow(128) // Pre-allocate buffer
    content.WriteString(m.fetching.stages[m.fetching.currentStage])
    content.WriteString("\n\n")
    content.WriteString(m.fetching.progress.View())
    content.WriteString("\n")
    return lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding).Render(content.String())
}

// commentsView renders the comments view
func (m model) commentsView() string {
	var content strings.Builder
	content.Grow(m.screenWidth * 10)
	content.WriteString(m.commentsViewport.View())

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Scroll"]+" • "+ keyNewComment +": "+Strings["AddComment"]+" • "+ keyExitKeysStr +": "+Strings["Back"]))

    return m.centralLayout(content.String(), footer)
}

// configListView renders the config list view
func (m model) configListView() string {
	return m.configList.list.View()
}

// configInputView renders the config input view
func (m model) configInputView() string {
	var content strings.Builder
	content.Grow(256) // Pre-allocate buffer
	content.WriteString(Strings["Edit"])
	content.WriteString(m.configInput.key)
	if m.configInput.key == "language" {
	    content.WriteString("\n"+ lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Only in this format: de-DE"))
	    content.WriteString("\n"+ lipgloss.NewStyle().Render("You can choose between these languages: " + strings.Join(allowedLanguages, ", ")))
	}
	content.WriteString("\n\n")
	content.WriteString(m.configInput.input.View())
	content.WriteString("\n\n")

    // Render save button
    if m.configInput.focusSave {
        content.WriteString(focusedButton)
    } else {
        content.WriteString(blurredButton)
    }

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+" • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+"  • esc: "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

// taskStatusView
func (m model) taskStatusView() string {
    var statuses = []string{
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
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseTaskStatus"])
	content.WriteString(Strings["SelectStatus"])

	for i, status := range statuses {
		if m.jiraStatusInput.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(status)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+" • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+"  • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

// addCommentView renders the Add Comment input view
func (m model) addCommentView() string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(Strings["AddCommentHeader"])
    content.WriteString(m.addCommentInput.input.View())
    content.WriteString("\n\n")

    // Render send button
    if m.addCommentInput.focusSend {
        content.WriteString(focusedButton)
    } else {
        content.WriteString(blurredButton)
    }

    // Footer
    parts := strings.Split(keyExitKeysStr, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Navigate"]+" • "+ keyMap[keyEnter] +": "+ Strings["Send"] +" • "+ parts[1] +": "+Strings["Cancel"]))

    return m.centralLayout(content.String(), footer)
}

// EnterCommitMessage renders the Enter Commit Message view
func (m model) enterCommitMessage() string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(Strings["AddCommitHeader"])
    content.WriteString("\n"+ lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(Strings["CommitNotice"]))
    content.WriteString("\n" + m.commitGitMessage.input.View())
    content.WriteString("\n\n")

    // Render send button
    if m.commitGitMessage.focusSend {
        content.WriteString(focusedButtonGit)
    } else {
        content.WriteString(blurredButtonGit)
    }

    // Footer
    parts := strings.Split(keyExitKeysStr, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Navigate"]+" • "+ keyMap[keyEnter] +": Copy Full Command • "+ parts[1] +": "+Strings["Cancel"]))

    return m.centralLayout(content.String(), footer)
}

func (m model) issueGitLocation() string {
    var pos = []string{
    	Strings["left"],
    	Strings["right"],
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseIssueGitLocation"])
	content.WriteString(Strings["SelectLocation"])

	for i, position := range pos {
		if m.gitIssueLoc.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(position)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+" • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+"  • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

func (m model) issueGitStyle() string {
    var styles = []string{
    	style1,
    	style2,
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseIssueGitStyle"])
	content.WriteString(Strings["SelectStyle"])

	for i, style := range styles {
		if m.gitIssueStyle.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(style)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+"  • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+" • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

func (m model) taskSearchView() string {
	var content strings.Builder
	content.Grow(512)

	// Render the search input
	content.WriteString("Search: \n")
	content.WriteString(m.taskSearchInput.View())
	content.WriteString("\n\n")

	// Render the filtered table
	content.WriteString(tableBaseStyle.Render(m.tasksTable.View()))
	content.WriteString("\n")

	// Render help
	footer := lipgloss.NewStyle().Faint(true).Render("esc: "+Strings["Cancel"]+" • " + keyMap[keyEnter] + ": "+Strings["Confirm"]+" & "+Strings["Cancel"])
	content.WriteString(footer)

	return content.String()
}