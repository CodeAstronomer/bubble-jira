package ui

import (
	"fmt"
    "os"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func getTerminalSize() (int, int) {
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        width = 80
    }
    return width, height
}

const (
	MenuViewTasksTitle  = "View Tasks"
	MenuSettingsTitle   = "Settings"
	MenuQuitTitle       = "Quit"
	MenuConfigTitle     = "Edit Config"
	MenuLicenceTitle    = "View Licence"
	MenuBackTitle       = "Back"
	ContextViewComments = "View Comments"
)

// Pre-computed style definitions
var (
    width, height      = getTerminalSize()
    topBottomPadding   = 1
    leftRightPadding   = 2
    terminalWidth      = (width-(leftRightPadding*leftRightPadding))
    terminalHeight      = (height-(topBottomPadding*topBottomPadding))
	menuStyle          = lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding)
	tasksStyle         = lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding)
	fetchingStyle      = lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding)
	configStyle        = lipgloss.NewStyle().Padding(topBottomPadding, leftRightPadding)
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	focusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle            = lipgloss.NewStyle()
	commentsStyle      = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("62")).
    Padding(0)
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	authorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	timeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
	helpStyle      = blurredStyle
	tableBaseStyle = lipgloss.NewStyle().Width(terminalWidth).
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240"))
	headerStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240")).
    BorderBottom(true).
    Bold(false)
	selectedStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("57")).
    Bold(false)
)

// Button styles
var (
	focusedButton = focusedStyle.Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
)

// Symbols
var (
	checkMark = successStyle.Render("✓")
)