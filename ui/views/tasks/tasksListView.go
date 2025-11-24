package tasks

import (
    "strings"
    "bubble-jira/ui/types"
	"github.com/charmbracelet/lipgloss"
)

func TasksTableView(data types.TasksListData) string {
	var content strings.Builder
	content.Grow(512)
	content.WriteString(data.TableBaseStyle.Render(data.TasksTable.View()))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(data.Keys.Up+"/"+ data.Keys.Down +": "+data.Strings["Navigate"]+" • ␣: "+data.Strings["Comments"]+"/"+data.Strings["Description"]+" • " + data.Keys.Enter +": "+data.Strings["Select"]+" • "+data.Keys.Search+": "+data.Strings["Search"]+" • "+ data.Keys.Exit +": "+data.Strings["Back"]))
	return content.String()
}