package views

import "bubble-jira/ui/types"

func ConfigListView(data types.ConfigListEditor) string {
	return data.List.View()
}
