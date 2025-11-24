package types

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type TaskSearchViewData struct {
	TaskSearchInput textinput.Model
	TasksTable      table.Model
	TableBaseStyle  lipgloss.Style
	Strings         map[string]string
	KeyMap          map[string]string
	KeyEnter        string
}
