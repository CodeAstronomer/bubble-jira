package comments

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func AddCommentView(data types.AddCommentViewData) string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(data.Strings["AddCommentHeader"])
    content.WriteString(data.AddCommentInputView)
    content.WriteString("\n\n")

    // Render send button
    if data.AddCommentInput {
        content.WriteString(data.FocusedButton)
    } else {
        content.WriteString(data.BlurredButton)
    }

    // Footer
    parts := strings.Split(data.Keys.Exit, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(data.Keys.Up+"/"+ data.Keys.Down +": "+data.Strings["Navigate"]+" • "+ data.Keys.Enter +": "+ data.Strings["Send"] +" • "+ parts[1] +": "+data.Strings["Cancel"]))

    return lipgloss.NewStyle().
        Width(data.ScreenWidth).
        Height(data.ScreenHeight).
        Render(content.String() + "\n\n" + footer)
}
