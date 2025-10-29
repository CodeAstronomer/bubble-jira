
package ui

import (
    "strings"
	"github.com/charmbracelet/lipgloss"
)

// licenceView renders the licence view
func (m model) licenceView() string {
    if m.licenceLoading {
        return "Loading licence..."
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
    footer := lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": scroll • "+ keyMap[keyFastUp]+"/"+ keyMap[keyFastDown] +": fast-scroll • "+ keyExitKeysStr +": back")

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
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": navigate • "+ keyMap[keyEnter] +": select • "+ keyExitKeysStr +": back"))
	return content.String()
}

// taskContextView renders the task context menu view
func (m model) taskContextView() string {
	return m.taskContextMenu.View()
}

// fetchingView renders the fetching view
func (m model) fetchingView() string {
	if m.fetching.error != "" {
		return lipgloss.NewStyle().Padding(1, 2).Render(
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
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": scroll • "+ keyExitKeysStr +": back"))
	return content.String()
}

// configListView renders the config list view
func (m model) configListView() string {
	return m.configList.list.View()
}

// configInputView renders the config input view
func (m model) configInputView() string {
	var content strings.Builder
	content.Grow(256) // Pre-allocate buffer
	content.WriteString("Editing: ")
	content.WriteString(m.configInput.key)
	content.WriteString("\n\n")
	content.WriteString(m.configInput.input.View())
	content.WriteString("\n\n")

    // Render save button
    if m.configInput.focusSave {
        content.WriteString(focusedButton)
    } else {
        content.WriteString(blurredButton)
    }
    content.WriteString("\n\n")
    content.WriteString(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": navigate • "+ keyMap[keyEnter] +": confirm • "+ keyExitKeysStr +": cancel"))


	return content.String()
}

// commitInputViewGit renders the git commit input screen
func (m model) commitInputViewGit() string {
    var content strings.Builder
    content.Grow(256)

    content.WriteString("🪶 Git Commit\n\n")
    content.WriteString("Issue: ")
    content.WriteString(m.commitInput.key)
    content.WriteString("\nLeave empty to use the task title as the commit message.")
    content.WriteString("\n\nMessage:\n")

    m.commitInput.input.Prompt = ""
    content.WriteString(m.commitInput.input.View())
    content.WriteString("\n\n")

    if m.commitInput.focusSave {
        content.WriteString(focusedButtonGit)
    } else {
        content.WriteString(blurredButtonGit)
    }

    content.WriteString("\n\n")
    content.WriteString(
        lipgloss.NewStyle().Faint(true).Render(
            keyMap[keyUp] + "/" + keyMap[keyDown] + ": navigate • " +
                keyMap[keyEnter] + ": confirm • " + keyExitKeysStr + ": cancel",
        ),
    )

    return content.String()
}