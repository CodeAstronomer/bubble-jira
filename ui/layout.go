package ui

import (
    "fmt"
    "strings"
	"github.com/charmbracelet/lipgloss"
)

func repeatNewline(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func (m model) centralLayout(content, footer string) string {
    contentHeight := lipgloss.Height(content)
    footerHeight := lipgloss.Height(footer)

    // Nutze m.screenHeight statt der globalen terminalHeight
    emptyLines := m.screenHeight - contentHeight - footerHeight
    if emptyLines < 0 {
        emptyLines = 0
    }

    finalContent := fmt.Sprintf("%s%s\n%s", content, repeatNewline(emptyLines), footer)

    // Nutze m.screenWidth statt der globalen terminalWidth
    return lipgloss.NewStyle().
        // Padding hier entfernen, da es in den Sub-Styles (menuStyle etc.) sein sollte
        Width(m.screenWidth).
        Height(m.screenHeight).
        Render(finalContent)
}