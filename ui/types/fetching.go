package types

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
)

// FetchingModel represents the fetching state UI
type FetchingModel struct {
	Spinner      spinner.Model
	Progress     progress.Model
	Stages       []string
	CurrentStage int
	FromStatus   bool
	Status       string
	Error        string
	Done         bool
	Width        int
}
