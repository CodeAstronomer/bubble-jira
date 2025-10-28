package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// licenceView renders the licence view
func (m model) licenceView() string {
	if m.licenceLoading {
		return "Loading licence...\n\nPress 'q' to go back"
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(
		m.licenceContent + "\n\nPress 'q' to go back",
	)
}

// tasksTableView renders the tasks table view
func (m model) tasksTableView() string {
	return m.tasksTable.View()
}

// taskContextView renders the task context menu view
func (m model) taskContextView() string {
	return m.taskContextMenu.View()
}

// fetchingView renders the fetching view
func (m model) fetchingView() string {
	if m.fetching.error != "" {
		return lipgloss.NewStyle().Padding(1, 2).Render(
			"Error: " + m.fetching.error + "\n\nPress 'q' to go back",
		)
	}

	content := strings.Builder{}
	content.WriteString(m.fetching.progress.View() + "\n")
	content.WriteString(m.fetching.spinner.View() + " " + m.fetching.status)
	return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}

// commentsView renders the comments view
func (m model) commentsView() string {
	return m.commentsViewport.View()
}

// configListView renders the config list view
func (m model) configListView() string {
	return m.configList.list.View()
}

// configInputView renders the config input view
func (m model) configInputView() string {
	content := strings.Builder{}
	content.WriteString("Editing: " + m.configInput.key + "\n\n")
	content.WriteString(m.configInput.input.View() + "\n\n")

	if m.configInput.focusSave {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("[ SAVE ]") + " ")
		content.WriteString("[ CANCEL ]\n")
	} else {
		content.WriteString("[ SAVE ] ")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("[ CANCEL ]") + "\n")
	}

	content.WriteString("\nPress 'tab' to switch focus, 'enter' to confirm, 'esc' to cancel")
	return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}