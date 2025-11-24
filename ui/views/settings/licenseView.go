package settings

import (
	"strings"
	"bubble-jira/ui/types"

	"github.com/charmbracelet/lipgloss"
)

func LicenseView(data types.LicenseViewData) string {
	if data.IsLoading {
		return data.Strings["LoadLicense"]
	}

	lines := strings.Split(data.Content, "\n")

	visibleLines := data.ScreenHeight - 2
	if data.Offset > len(lines)-visibleLines {
		data.Offset = max(0, len(lines)-visibleLines)
	}
	if data.Offset < 0 {
		data.Offset = 0
	}

	end := data.Offset + visibleLines
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[data.Offset:end]

	content := strings.Join(visible, "\n")
	footer := lipgloss.NewStyle().Faint(true).Render(data.Keys.Up+"/"+data.Keys.Down+": "+data.Strings["Scroll"]+" • "+data.Keys.FastUp+"/"+data.Keys.FastDown+": "+data.Strings["FastScroll"]+" • "+data.Keys.Exit+": "+data.Strings["Back"])

	return lipgloss.NewStyle().
		Width(data.ScreenWidth).
		Height(data.ScreenHeight).
		Render(content + "\n\n" + footer)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}