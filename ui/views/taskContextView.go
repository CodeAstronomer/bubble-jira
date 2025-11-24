package views

import "bubble-jira/ui/types"

func TaskContextView(data types.TaskContextViewData) string {
	return data.Menu.View()
}
