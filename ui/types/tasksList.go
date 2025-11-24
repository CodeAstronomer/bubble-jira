package types

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type TasksListData struct {
	Content        string
	Keys           KeyData
	Strings        map[string]string
	TasksTable     table.Model
	TableBaseStyle lipgloss.Style
}