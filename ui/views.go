package ui

import (
    "strings"
	"github.com/charmbracelet/lipgloss"
)


// configInputView renders the config input view
func (m model) configInputView() string {
	var content strings.Builder
	content.Grow(256) // Pre-allocate buffer
	content.WriteString(Strings["Edit"])
	content.WriteString(m.configInput.key)
	content.WriteString("\n\n")
	content.WriteString(m.configInput.input.View())
	content.WriteString("\n\n")

    // Render save button
    if m.configInput.focusSave {
        content.WriteString(focusedButton)
    } else {
        content.WriteString(blurredButton)
    }

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+" • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+"  • esc: "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

func (m model) issueGitLocation() string {
    var pos = []string{
    	Strings["left"],
    	Strings["right"],
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseIssueGitLocation"])
	content.WriteString(Strings["SelectLocation"])

	for i, position := range pos {
		if m.gitIssueLoc.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(position)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+" • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+"  • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

func (m model) issueGitStyle() string {
    var styles = []string{
    	style1,
    	style2,
    }
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseIssueGitStyle"])
	content.WriteString(Strings["SelectStyle"])

	for i, style := range styles {
		if m.gitIssueStyle.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(style)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+"  • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+" • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}

func (m model) selectLanguage() string {
	var content strings.Builder
	content.Grow(256)
	content.WriteString(Strings["ChooseLanguages"])

	for i, chooseLanguages := range allowedLanguages {
		if m.ChooseLanguage.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(chooseLanguages)
		content.WriteString("\n")
	}

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(
        keyMap[keyUp]+"/"+keyMap[keyDown]+": "+Strings["Navigate"]+"  • "+
            keyMap[keyEnter]+": "+Strings["Confirm"]+" • "+
            keyExitKeysStr+": "+Strings["Cancel"],
    )

    return m.centralLayout(content.String(), footer)
}
