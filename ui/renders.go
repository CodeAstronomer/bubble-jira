
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
	var content strings.Builder
    content.Grow(len(m.licenceContent) + 100)
    content.WriteString(m.licenceContent)
    content.WriteString("\n\n")
    content.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓: scroll • q/esc: back"))
    return content.String()
}

// tasksTableView renders the tasks table view
func (m model) tasksTableView() string {
	var content strings.Builder
	content.Grow(512)
	content.WriteString(tableBaseStyle.Render(m.tasksTable.View()))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • enter: select • q: back"))
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
    return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}

// commentsView renders the comments view
func (m model) commentsView() string {
	var content strings.Builder
	content.Grow(m.screenWidth * 10)
	content.WriteString(m.commentsViewport.View())
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓: scroll • q/esc: back"))
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

	return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}