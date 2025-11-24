package views

import (
	"bubble-jira/ui/types"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func TaskSearchView(data types.TaskSearchViewData) string {
	var content strings.Builder
	content.Grow(512)

	// Render the search input
	content.WriteString("Search: \n")
	content.WriteString(data.TaskSearchInput.View())
	content.WriteString("\n\n")

	// Render the filtered table
	content.WriteString(data.TableBaseStyle.Render(data.TasksTable.View()))
	content.WriteString("\n")

	// Render help
	footer := lipgloss.NewStyle().Faint(true).Render("esc: "+data.Strings["Cancel"]+" • " + data.KeyMap[data.KeyEnter] + ": "+data.Strings["Confirm"]+" & "+data.Strings["Cancel"])
	content.WriteString(footer)

	return content.String()
}
