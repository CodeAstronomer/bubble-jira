package settings

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func IssueGitLocation(data types.IssueGitLocationData) string {
    var pos = []string{
    	data.Strings["left"],
    	data.Strings["right"],
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(data.Strings["ChooseIssueGitLocation"])
	content.WriteString(data.Strings["SelectLocation"])

	for i, position := range pos {
		if data.GitIssueLocCursor == i {
			content.WriteString(data.Strings["posTrue"])
		} else {
			content.WriteString(data.Strings["posFalse"])
		}
		content.WriteString(position)
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