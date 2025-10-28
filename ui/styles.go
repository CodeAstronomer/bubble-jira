package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

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
	menuStyle          = lipgloss.NewStyle().Padding(1, 2)
	tasksStyle         = lipgloss.NewStyle().Padding(1, 2)
	fetchingStyle      = lipgloss.NewStyle().Padding(2, 4)
	configStyle        = lipgloss.NewStyle().Padding(1, 2)
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
	tableBaseStyle = lipgloss.NewStyle().
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