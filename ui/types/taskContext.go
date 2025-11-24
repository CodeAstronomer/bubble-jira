package types

import "github.com/charmbracelet/bubbles/list"

type TaskContextViewData struct {
	Menu list.Model
}

// ContextMenuItem represents a context menu item
type ContextMenuItem struct {
	title string
}

func NewContextMenuItem(title string) ContextMenuItem {
	return ContextMenuItem{title: title}
}

func (cm ContextMenuItem) Title() string       { return cm.title }
func (cm ContextMenuItem) Description() string { return "" }
func (cm ContextMenuItem) FilterValue() string { return cm.title }
