package comments

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func CommentsView(data types.CommentViewData) string {
	var content strings.Builder
	content.Grow(data.ScreenWidth * 10)
	content.WriteString(data.CommentsViewPort)

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(data.Keys.Up+"/"+ data.Keys.Down +": "+data.Strings["Scroll"]+" • "+ data.Keys.Comment +": "+data.Strings["AddComment"]+" • "+ data.Keys.Exit +": "+data.Strings["Back"]))

    return lipgloss.NewStyle().
        Width(data.ScreenWidth).
        Height(data.ScreenHeight).
        Render(content.String() + "\n\n" + footer)
}