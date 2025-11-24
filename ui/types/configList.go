package types

import "github.com/charmbracelet/bubbles/list"

type ConfigListEditor struct {
	Fields []ConfigField
	List   list.Model
}

type ConfigField struct {
	Key        string
	DisplayKey string
	Value      string
}

func (cf ConfigField) Title() string       { return cf.DisplayKey }
func (cf ConfigField) Description() string { return cf.Value }
func (cf ConfigField) FilterValue() string { return cf.DisplayKey }
