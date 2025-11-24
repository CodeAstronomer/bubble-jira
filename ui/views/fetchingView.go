package views

import (
	"bubble-jira/ui/types"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const (
    topBottomPadding = 1
    leftRightPadding = 2
)

func FetchingView(data types.FetchingModel) string {
	if data.Error != "" {
		return lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding).Render(
			"Error: " + data.Error,
		)
	}

	var content strings.Builder
	content.Grow(128) // Pre-allocate buffer
	content.WriteString(data.Stages[data.CurrentStage])
	content.WriteString("\n\n")
	content.WriteString(data.Progress.View())
	content.WriteString("\n")
	return lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding).Render(content.String())
}
