package settings

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func SelectLanguage(data types.SelectLanguageData) string {
	var content strings.Builder
	content.Grow(256)
	content.WriteString(data.Strings["ChooseLanguages"])

	for i, chooseLanguages := range data.AllowedLanguages {
		if data.ChooseLanguageCursor == i {
			content.WriteString(data.Strings["posTrue"])
		} else {
			content.WriteString(data.Strings["posFalse"])
		}
		content.WriteString(chooseLanguages)
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