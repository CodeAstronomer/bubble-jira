package settings

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func IssueGitStyle(data types.IssueGitStyleData) string {
    var styles = []string{
    	data.Style1,
    	data.Style2,
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(data.Strings["ChooseIssueGitStyle"])
	content.WriteString(data.Strings["SelectStyle"])

	for i, style := range styles {
		if data.GitIssueStyleCursor == i {
			content.WriteString(data.Strings["posTrue"])
		} else {
			content.WriteString(data.Strings["posFalse"])
		}
		content.WriteString(style)
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