package settings

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func ConfigInputView(data types.ConfigInputViewData) string {
	var content strings.Builder
	content.Grow(256) // Pre-allocate buffer
	content.WriteString(data.Strings["Edit"])
	content.WriteString(data.ConfigInputKey)
	content.WriteString("\n\n")
	content.WriteString(data.ConfigInputView)
	content.WriteString("\n\n")

    // Render save button
    if data.ConfigInput {
        content.WriteString(data.FocusedButton)
    } else {
        content.WriteString(data.BlurredButton)
    }

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        data.Keys.Up+"/"+data.Keys.Down+": "+data.Strings["Navigate"]+" • "+
            data.Keys.Enter+": "+data.Strings["Confirm"]+"  • esc: "+data.Strings["Cancel"],
    )

    return lipgloss.NewStyle().
        Width(data.ScreenWidth).
        Height(data.ScreenHeight).
        Render(content.String() + "\n\n" + footer)
}