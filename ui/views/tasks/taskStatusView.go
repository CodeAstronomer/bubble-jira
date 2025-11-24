package tasks

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

// taskStatusView
func TaskStatusView(data types.TaskStatusViewData) string {
	var content strings.Builder
	content.Grow(256)

	content.WriteString(data.Strings["ChooseTaskStatus"])
	content.WriteString(data.Strings["SelectStatus"])

	for i, status := range data.Statuses {
		if data.JiraStatusInput == i {
			content.WriteString(data.Strings["posTrue"])
		} else {
			content.WriteString(data.Strings["posFalse"])
		}
		content.WriteString(status)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        data.Keys.Up+"/"+data.Keys.Down+": "+data.Strings["Navigate"]+" • "+
            data.Keys.Enter+": "+data.Strings["Confirm"]+"  • "+
            data.Keys.Exit+": "+data.Strings["Cancel"],
    )

    return lipgloss.NewStyle().
        Width(data.ScreenWidth).
        Height(data.ScreenHeight).
        Render(content.String() + "\n\n" + footer)
}