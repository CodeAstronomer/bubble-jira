package ui

import (
    "strings"
	"github.com/charmbracelet/lipgloss"
)

// commentsView renders the comments view
func (m model) commentsView() string {
	var content strings.Builder
	content.Grow(m.screenWidth * 10)
	content.WriteString(m.commentsViewport.View())

    // Footer
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Scroll"]+" • "+ keyNewComment +": "+Strings["AddComment"]+" • "+ keyExitKeysStr +": "+Strings["Back"]))

    return m.centralLayout(content.String(), footer)
}



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

// taskStatusView
func (m model) taskStatusView() string {
	var content strings.Builder
	content.Grow(256)

	content.WriteString(Strings["ChooseTaskStatus"])
	content.WriteString(Strings["SelectStatus"])

	for i, status := range statuses {
		if m.jiraStatusInput.cursor == i {
			content.WriteString(Strings["posTrue"])
		} else {
			content.WriteString(Strings["posFalse"])
		}
		content.WriteString(status)
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

// addCommentView renders the Add Comment input view
func (m model) addCommentView() string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(Strings["AddCommentHeader"])
    content.WriteString(m.addCommentInput.input.View())
    content.WriteString("\n\n")

    // Render send button
    if m.addCommentInput.focusSend {
        content.WriteString(focusedButton)
    } else {
        content.WriteString(blurredButton)
    }

    // Footer
    parts := strings.Split(keyExitKeysStr, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Navigate"]+" • "+ keyMap[keyEnter] +": "+ Strings["Send"] +" • "+ parts[1] +": "+Strings["Cancel"]))

    return m.centralLayout(content.String(), footer)
}

// EnterCommitMessage renders the Enter Commit Message view
func (m model) enterCommitMessage() string {
    var content strings.Builder
    content.Grow(512)

    content.WriteString(Strings["AddCommitHeader"])
    content.WriteString("\n"+ lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(Strings["CommitNotice"]))
    content.WriteString("\n" + m.commitGitMessage.input.View())
    content.WriteString("\n\n")

    // Render send button
    if m.commitGitMessage.focusSend {
        content.WriteString(focusedButtonGit)
    } else {
        content.WriteString(blurredButtonGit)
    }

    // Footer
    parts := strings.Split(keyExitKeysStr, "/")
    footer := lipgloss.NewStyle().Faint(true).Render(lipgloss.NewStyle().Faint(true).Render(keyMap[keyUp]+"/"+ keyMap[keyDown] +": "+Strings["Navigate"]+" • "+ keyMap[keyEnter] +": Copy Full Command • "+ parts[1] +": "+Strings["Cancel"]))

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
