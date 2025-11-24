package types

type TaskStatusViewData struct {
	ScreenHeight        int
	ScreenWidth         int
	Keys                KeyData
	Strings             map[string]string
	Statuses            []string
	JiraStatusInput     any
}