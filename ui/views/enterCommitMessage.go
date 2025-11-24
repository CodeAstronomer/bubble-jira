package views

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func EnterCommitMessage(data types.EnterCommitMessageData) string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(data.Strings["AddCommitHeader"])
    content.WriteString("\n"+ lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(data.Strings["CommitNotice"]))
    content.WriteString("\n" + data.CommitGitMessageView)
    content.WriteString("\n\n")

    // Render send button
    if data.CommitGitMessage {
        content.WriteString(data.FocusedButtonGit)
    } else {
        content.WriteString(data.BlurredButtonGit)
    }

    // Footer
    parts := strings.Split(data.Keys.Exit, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(data.Keys.Up+"/"+ data.Keys.Down +": "+data.Strings["Navigate"]+" • "+ data.Keys.Enter +": Copy Full Command • "+ parts[1] +": "+data.Strings["Cancel"]))

    return lipgloss.NewStyle().
        Width(data.ScreenWidth).
        Height(data.ScreenHeight).
        Render(content.String() + "\n\n" + footer)
}